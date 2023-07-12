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
	"github.com/interline-io/transitland-lib/dmfr"
	"github.com/interline-io/transitland-lib/dmfr/fetch"
	"github.com/interline-io/transitland-lib/log"
	"github.com/interline-io/transitland-lib/rt/pb"
	"github.com/interline-io/transitland-lib/tl/tt"
	"github.com/interline-io/transitland-lib/tldb"
	"github.com/interline-io/transitland-server/auth/authn"
	"github.com/interline-io/transitland-server/config"
	"github.com/interline-io/transitland-server/internal/dbutil"
	"github.com/interline-io/transitland-server/internal/generated/azpb"
	"github.com/interline-io/transitland-server/model"
	"github.com/jmoiron/sqlx"
	"google.golang.org/protobuf/proto"
)

func StaticFetch(ctx context.Context, cfg config.Config, dbf model.Finder, feedId string, feedSrc io.Reader, feedUrl string, user authn.User, checker model.Checker) (*model.FeedVersionFetchResult, error) {
	urlType := "static_current"
	feed, err := fetchCheckFeed(ctx, dbf, checker, feedId, urlType, feedUrl)
	if err != nil {
		return nil, err
	}
	if feed == nil {
		return nil, nil
	}

	// Prepare
	fetchOpts := fetch.Options{
		FeedID:    feed.ID,
		URLType:   urlType,
		FeedURL:   feedUrl,
		Storage:   cfg.Storage,
		Secrets:   cfg.Secrets,
		FetchedAt: time.Now().In(time.UTC),
	}
	if user != nil {
		fetchOpts.CreatedBy = tt.NewString(user.Name())
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

func RTFetch(ctx context.Context, cfg config.Config, dbf model.Finder, rtf model.RTFinder, target string, feedId string, feedUrl string, urlType string, checker model.Checker) error {
	feed, err := fetchCheckFeed(ctx, dbf, checker, feedId, urlType, feedUrl)
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
	if err := tldb.NewPostgresAdapterFromDBX(dbf.DBX()).Tx(func(atx tldb.Adapter) error {
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
	return rtf.AddData(key, rtdata)
}

func CheckFetchWait(ctx context.Context, db sqlx.Ext, feedId int, urlType string, fetchWait time.Duration) (bool, error) {
	// Check if minimum fetch time has elapsed
	now := time.Now().In(time.UTC)
	q := sq.
		Select("feed_fetches.*").
		From("feed_fetches").
		Where(sq.Eq{"feed_id": feedId}).
		Where(sq.Eq{"url_type": urlType}).
		OrderBy("fetched_at desc").
		Limit(10)
	var lastFetches []dmfr.FeedFetch
	if err := dbutil.Select(ctx, db, q, &lastFetches); err != nil {
		return false, err
	}
	if len(lastFetches) == 0 {
		return true, nil
	}

	// Get exponential backoff
	// Based on failures in the past 24 hours
	failures := 0
	lastFetchAgo := time.Duration(0)
	checkFailureWindow := time.Duration(60*60*24) * time.Second
	for _, fetch := range lastFetches {
		timeAgo := now.Sub(fetch.FetchedAt.Val.In(time.UTC))
		if lastFetchAgo == 0 || timeAgo < lastFetchAgo {
			lastFetchAgo = timeAgo
		}
		if fetch.Success {
			break
		}
		if timeAgo > checkFailureWindow {
			break
		}
		failures += 1
	}
	failureBackoff := time.Duration(math.Pow(4, float64(failures+1))) * time.Second
	if failureBackoff > checkFailureWindow {
		failureBackoff = checkFailureWindow
	}
	// fmt.Println("\tfailures:", failures, "failureBackoff:", failureBackoff)
	if failures > 0 && lastFetchAgo < failureBackoff {
		log.Trace().Int("feed_id", feedId).Str("url_type", urlType).Float64("failure_backoff", failureBackoff.Seconds()).Int("failures", failures).Float64("last_fetch_ago", lastFetchAgo.Seconds()).Msg("skipping - failure backoff")
		return false, nil
	}
	if lastFetchAgo < fetchWait {
		log.Trace().Int("feed_id", feedId).Str("url_type", urlType).Float64("fetch_wait", fetchWait.Seconds()).Float64("last_fetch_ago", lastFetchAgo.Seconds()).Msg("skipping - fetch wait")
		return false, nil
	}
	return true, nil
}

func fetchCheckFeed(ctx context.Context, dbf model.Finder, checker model.Checker, feedId string, urlType string, url string) (*model.Feed, error) {
	// Check feed exists
	feeds, err := dbf.FindFeeds(ctx, nil, nil, nil, nil, &model.FeedFilter{OnestopID: &feedId})
	if err != nil {
		return nil, err
	}
	if len(feeds) == 0 {
		return nil, errors.New("feed not found")
	}
	feed := feeds[0]

	// Check feed permissions
	if checker != nil {
		if check, err := checker.FeedPermissions(ctx, &azpb.FeedRequest{Id: int64(feed.ID)}); err != nil {
			return nil, err
		} else if !check.Actions.CanCreateFeedVersion {
			return nil, errors.New("unauthorized")
		}
	}

	// Re-check last fetched
	db := dbf.DBX()
	if ok, err := CheckFetchWait(ctx, db, feed.ID, urlType, time.Duration(feed.FetchWait.Val)*time.Second); err != nil {
		return nil, err
	} else if !ok {
		return nil, nil
	}
	return feed, nil
}
