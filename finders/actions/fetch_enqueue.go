package actions

import (
	"context"
	"errors"
	"math"
	"time"

	sq "github.com/Masterminds/squirrel"
	"github.com/interline-io/log"
	"github.com/interline-io/transitland-dbutil/dbutil"
	"github.com/interline-io/transitland-jobs/jobs"
	"github.com/interline-io/transitland-lib/tl"
	"github.com/interline-io/transitland-lib/tl/tt"
	"github.com/interline-io/transitland-mw/auth/authz"
	"github.com/jmoiron/sqlx"

	"github.com/interline-io/transitland-server/model"
)

func FetchEnqueue(ctx context.Context, feedIds []string, urlTypes []string, ignoreFetchWait bool) error {
	cfg := model.ForContext(ctx)
	db := cfg.Finder.DBX()
	now := time.Now().In(time.UTC)
	feeds, err := cfg.Finder.FindFeeds(ctx, nil, nil, nil, &model.FeedFilter{})
	if err != nil {
		return err
	}

	// Check onestop ids
	checkOsid := map[string]bool{}
	for _, v := range feedIds {
		checkOsid[v] = true
	}
	feedLookup := map[int]*model.Feed{}
	for _, feed := range feeds {
		if len(checkOsid) > 0 && !checkOsid[feed.FeedID] {
			continue
		}
		feedLookup[feed.ID] = feed
	}

	// Check url types
	staticKeys := []string{"static_current"}
	rtKeys := []string{"realtime_alerts", "realtime_trip_updates", "realtime_vehicle_positions"}
	gbfsKeys := []string{"gbfs_auto_discovery"}
	checkUrlTypes := map[string]bool{}
	for _, v := range urlTypes {
		checkUrlTypes[v] = true
	}
	if len(checkUrlTypes) == 0 {
		for _, a := range [][]string{staticKeys, rtKeys, gbfsKeys} {
			for _, v := range a {
				checkUrlTypes[v] = true
			}
		}
	}

	// Check for fetch wait times and error backoffs
	// This is structured to avoid bad N+1 queries
	feedsByUrlType := map[string][]CheckFetchWaitResult{}
	for urlType, ok := range checkUrlTypes {
		if !ok {
			continue
		}
		var feedIds []int
		for _, feed := range feedLookup {
			if getUrl(feed.URLs, urlType) != "" {
				feedIds = append(feedIds, feed.ID)
			}
		}

		feedChecks, err := CheckFetchWaitBatch(ctx, db, feedIds, urlType)
		if err != nil {
			return err
		}
		var feedsOk []CheckFetchWaitResult
		for _, check := range feedChecks {
			if check.OK() || ignoreFetchWait {
				feedsOk = append(feedsOk, check)
			}
		}
		feedsByUrlType[urlType] = feedsOk
	}

	// Static fetch
	var jj []jobs.Job
	for _, urlType := range staticKeys {
		for _, check := range feedsByUrlType[urlType] {
			feed, ok := feedLookup[check.ID]
			if !ok {
				continue
			}
			// TODO: Make this type safe, use StaticFetchWorker{} as job args
			jj = append(jj, jobs.Job{
				Queue:       "static-fetch",
				JobType:     "static-fetch",
				Unique:      true,
				JobDeadline: now.Add(check.Deadline()).Unix(),
				JobArgs: jobs.JobArgs{
					"feed_url":    getUrl(feed.URLs, urlType),
					"feed_id":     feed.FeedID,
					"fetch_epoch": 0,
				},
			})
		}
	}

	// RT fetch
	for _, urlType := range rtKeys {
		for _, check := range feedsByUrlType[urlType] {
			feed, ok := feedLookup[check.ID]
			if !ok {
				continue
			}
			// TODO: Make this type safe, use StaticFetchWorker{} as job args
			jj = append(jj, jobs.Job{
				Queue:       "rt-fetch",
				JobType:     "rt-fetch",
				Unique:      true,
				JobDeadline: now.Add(check.Deadline()).Unix(),
				JobArgs: jobs.JobArgs{
					"target":         feed.FeedID,
					"url":            getUrl(feed.URLs, urlType),
					"source_type":    urlType,
					"source_feed_id": feed.FeedID,
					"fetch_epoch":    0,
				},
			})
		}
	}

	// GBFS fetch
	for _, urlType := range gbfsKeys {
		for _, check := range feedsByUrlType[urlType] {
			feed, ok := feedLookup[check.ID]
			if !ok {
				continue
			}
			jj = append(jj, jobs.Job{
				Queue:       "gbfs-fetch",
				JobType:     "gbfs-fetch",
				Unique:      true,
				JobDeadline: now.Add(check.Deadline()).Unix(),
				JobArgs: jobs.JobArgs{
					"url":         getUrl(feed.URLs, urlType),
					"feed_id":     feed.FeedID,
					"fetch_epoch": 0,
				},
			})
		}
	}

	if jobQueue := cfg.JobQueue; jobQueue == nil {
		return errors.New("no job queue available")
	} else {
		if err := jobQueue.AddJobs(ctx, jj); err != nil {
			return err
		}
	}
	return nil
}

func getUrl(urls tl.FeedUrls, urlType string) string {
	url := ""
	switch urlType {
	case "static_current":
		url = urls.StaticCurrent
	case "realtime_alerts":
		url = urls.RealtimeAlerts
	case "realtime_trip_updates":
		url = urls.RealtimeTripUpdates
	case "realtime_vehicle_positions":
		url = urls.RealtimeVehiclePositions
	case "gbfs_auto_discovery":
		url = urls.GbfsAutoDiscovery
	}
	return url
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
		Time("last_fetched_at_plus_backoff", lastFetchedAt.Add(failureBackoff)).
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
