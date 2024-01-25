package actions

import (
	"context"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"math"
	"os"
	"time"

	sq "github.com/Masterminds/squirrel"
	"github.com/interline-io/log"
	"github.com/interline-io/transitland-lib/dmfr/fetch"
	"github.com/interline-io/transitland-lib/rt/pb"
	"github.com/interline-io/transitland-lib/tl/tt"
	"github.com/interline-io/transitland-lib/tldb"
	"github.com/interline-io/transitland-mw/auth/authn"
	"github.com/interline-io/transitland-mw/auth/authz"
	"github.com/interline-io/transitland-server/internal/dbutil"
	"github.com/interline-io/transitland-server/model"
	"github.com/jmoiron/sqlx"
	"google.golang.org/protobuf/proto"
)

func StaticFetch(ctx context.Context, feedId string, feedSrc io.Reader, feedUrl string) (*model.FeedVersionFetchResult, error) {
	cfg := model.ForContext(ctx)
	dbf := cfg.Finder

	urlType := "static_current"
	feed, err := fetchCheckFeed(ctx, feedId)
	if err != nil {
		return nil, err
	}
	if feed == nil {
		return nil, nil
	}

	// Prepare
	fetchOpts := fetch.Options{
		FeedID:        feed.ID,
		URLType:       urlType,
		FeedURL:       feedUrl,
		Storage:       cfg.Storage,
		Secrets:       cfg.Secrets,
		FetchedAt:     time.Now().In(time.UTC),
		AllowFTPFetch: true,
	}

	if user := authn.ForContext(ctx); user != nil {
		fetchOpts.CreatedBy = tt.NewString(user.ID())
	}

	// Allow a Reader
	if feedSrc != nil {
		tmpfile, err := ioutil.TempFile("", "validator-upload")
		if err != nil {
			return nil, err
		}
		io.Copy(tmpfile, feedSrc)
		tmpfile.Close()
		defer os.Remove(tmpfile.Name())
		fetchOpts.FeedURL = tmpfile.Name()
		fetchOpts.AllowLocalFetch = true
	}

	// Make request
	mr := model.FeedVersionFetchResult{}
	db := tldb.NewPostgresAdapterFromDBX(dbf.DBX())
	if err := db.Tx(func(atx tldb.Adapter) error {
		fv, fr, err := fetch.StaticFetch(atx, fetchOpts)
		if err != nil {
			return err
		}
		mr.FoundSHA1 = fr.Found
		if fr.FetchError == nil {
			mr.FeedVersion = &model.FeedVersion{FeedVersion: fv}
			mr.FetchError = nil
		} else {
			a := fr.FetchError.Error()
			mr.FetchError = &a
		}
		return nil
	}); err != nil {
		return nil, err
	}
	return &mr, nil
}

func RTFetch(ctx context.Context, target string, feedId string, feedUrl string, urlType string) error {
	cfg := model.ForContext(ctx)

	feed, err := fetchCheckFeed(ctx, feedId)
	if err != nil {
		return err
	}
	if feed == nil {
		return nil
	}

	// Prepare
	fetchOpts := fetch.Options{
		FeedID:    feed.ID,
		URLType:   urlType,
		FeedURL:   feedUrl,
		Storage:   cfg.RTStorage,
		Secrets:   cfg.Secrets,
		FetchedAt: time.Now().In(time.UTC),
	}

	// Make request
	var rtMsg *pb.FeedMessage
	var fetchErr error
	if err := tldb.NewPostgresAdapterFromDBX(cfg.Finder.DBX()).Tx(func(atx tldb.Adapter) error {
		m, fr, err := fetch.RTFetch(atx, fetchOpts)
		if err != nil {
			return err
		}
		rtMsg = m
		fetchErr = fr.FetchError
		return nil
	}); err != nil {
		return err
	}

	// Check result and cache
	if fetchErr != nil {
		return fetchErr
	}
	rtdata, err := proto.Marshal(rtMsg)
	if err != nil {
		return errors.New("invalid rt data")
	}
	key := fmt.Sprintf("rtdata:%s:%s", target, urlType)
	return cfg.RTFinder.AddData(key, rtdata)
}

type CheckFetchWaitResult struct {
	ID               int
	OnestopID        tt.String
	FetchWait        tt.Int
	LastFetchedAt    tt.Time
	URLType          string
	DefaultFetchWait int64
	Failures         int
	CheckedAt        time.Time
}

func (check *CheckFetchWaitResult) OK() bool {
	now := check.CheckedAt
	lastFetchedAt := check.LastFetchedAt.Val
	lastFetchAgo := now.Sub(lastFetchedAt)
	failureBackoff := time.Duration(0)
	if check.Failures > 0 {
		failureBackoff = time.Duration(math.Pow(4, float64(check.Failures+1))) * time.Second
		failureBackoffMax := (time.Duration(24*24*60) * time.Second)
		if failureBackoff.Seconds() > failureBackoffMax.Seconds() {
			failureBackoff = failureBackoffMax
		}
	}
	fetchWait := time.Duration(check.DefaultFetchWait) * time.Second
	if check.FetchWait.Valid {
		fetchWait = time.Duration(check.FetchWait.Val) * time.Second
	}

	logMsg := log.Trace().
		Int("feed_id", check.ID).
		Str("onestop_id", check.OnestopID.Val).
		Str("url_type", check.URLType).
		Float64("failure_backoff", failureBackoff.Seconds()).
		Int("failures", check.Failures).
		Str("last_fetched_at", lastFetchedAt.String()).
		Float64("last_fetch_ago", lastFetchAgo.Seconds()).
		Float64("fetch_wait", fetchWait.Seconds())

	if check.Failures > 0 && failureBackoff > 0 && now.Before(lastFetchedAt.Add(failureBackoff)) {
		logMsg.Msg("fetch wait: skipping, failure backoff")
		return false
	}
	if check.LastFetchedAt.Valid && fetchWait > 0 && now.Before(lastFetchedAt.Add(fetchWait)) {
		logMsg.Msg("fetch wait: skipping, too soon")
		return false
	}
	logMsg.Msg("fetch wait: ok")
	return true
}

func (check *CheckFetchWaitResult) Deadline() time.Duration {
	// TODO: Consider deadline based on fetchWait/failureBackoff
	deadline := time.Duration(60*60) * time.Second
	return deadline
}

func CheckFetchWaitBatch(ctx context.Context, db sqlx.Ext, feedIds []int, urlType string) ([]CheckFetchWaitResult, error) {
	now := time.Now().In(time.UTC)
	defaultFetchWait := int64(3600)
	defaultFetchWaitUrlType := map[string]int64{
		"static_current":             3600,
		"gbfs_auto_discovery":        600,
		"realtime_alerts":            60,
		"realtime_trip_updates":      60,
		"realtime_vehicle_positions": 60,
	}
	checks := map[int]CheckFetchWaitResult{}
	for _, chunk := range chunkBy(feedIds, 1000) {
		q := sq.Select("cf.id", "cf.onestop_id", "fs.fetch_wait", "ff.fetched_at", "ff.success").
			From("current_feeds cf").
			Join("feed_states fs on fs.feed_id = cf.id").
			JoinClause(`left join lateral (select fetched_at, success from feed_fetches ff where ff.feed_id = cf.id and ff.url_type = ? and ff.fetched_at >= (NOW() - INTERVAL '1 DAY') order by fetched_at desc limit 8) ff on true`, urlType).
			Where(sq.Eq{"cf.id": chunk})
		var lastFetches []struct {
			ID        int
			OnestopID tt.String
			FetchWait tt.Int
			FetchedAt tt.Time
			Success   tt.Bool
		}
		if err := dbutil.Select(ctx, db, q, &lastFetches); err != nil {
			return nil, err
		}
		for _, fetch := range lastFetches {
			a := checks[fetch.ID]
			a.CheckedAt = now
			a.ID = fetch.ID
			a.OnestopID = fetch.OnestopID
			a.DefaultFetchWait = defaultFetchWait
			if v, ok := defaultFetchWaitUrlType[urlType]; ok {
				a.DefaultFetchWait = v
			}
			if fetch.FetchWait.Valid {
				a.FetchWait = fetch.FetchWait
			}
			if fetch.Success.Valid && !fetch.Success.Val {
				a.Failures += 1
			}
			if fetch.FetchedAt.Valid && fetch.FetchedAt.Val.After(a.LastFetchedAt.Val) {
				a.LastFetchedAt = fetch.FetchedAt
			}
			checks[fetch.ID] = a
		}
	}
	var ret []CheckFetchWaitResult
	for _, check := range checks {
		ret = append(ret, check)
	}
	return ret, nil
}

func chunkBy[T any](items []T, chunkSize int) (chunks [][]T) {
	for chunkSize < len(items) {
		items, chunks = items[chunkSize:], append(chunks, items[0:chunkSize:chunkSize])
	}
	return append(chunks, items)
}

func fetchCheckFeed(ctx context.Context, feedId string) (*model.Feed, error) {
	// Check feed exists
	cfg := model.ForContext(ctx)
	feeds, err := cfg.Finder.FindFeeds(ctx, nil, nil, nil, &model.FeedFilter{OnestopID: &feedId})
	if err != nil {
		return nil, err
	}
	if len(feeds) == 0 {
		return nil, errors.New("feed not found")
	}
	feed := feeds[0]

	// Check feed permissions
	if checker := cfg.Checker; checker == nil {
		// pass
	} else if check, err := checker.FeedPermissions(ctx, &authz.FeedRequest{Id: int64(feed.ID)}); err != nil {
		return nil, err
	} else if !check.Actions.CanCreateFeedVersion {
		return nil, errors.New("unauthorized")
	}
	return feed, nil
}
