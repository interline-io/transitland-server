package workers

import (
	"context"
	"time"

	"github.com/interline-io/transitland-lib/tl"
	"github.com/interline-io/transitland-server/actions"
	"github.com/interline-io/transitland-server/jobs"
	"github.com/interline-io/transitland-server/model"
)

type FetchEnqueueWorker struct {
	URLTypes []string `json:"url_types"`
	FeedIDs  []string `json:"feed_ids"`
}

func (w *FetchEnqueueWorker) Run(ctx context.Context, job jobs.Job) error {
	cfg := model.ForContext(ctx)
	db := cfg.Finder.DBX()
	opts := job.Opts
	now := time.Now().In(time.UTC)
	feeds, err := cfg.Finder.FindFeeds(ctx, nil, nil, nil, &model.FeedFilter{})
	if err != nil {
		return err
	}

	// Check onestop ids
	checkOsid := map[string]bool{}
	for _, v := range w.FeedIDs {
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
	for _, v := range w.URLTypes {
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
	feedsByUrlType := map[string][]actions.CheckFetchWaitResult{}
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
		feedChecks, err := actions.CheckFetchWaitBatch(ctx, db, feedIds, urlType)
		if err != nil {
			return err
		}
		var feedsOk []actions.CheckFetchWaitResult
		for _, check := range feedChecks {
			if check.OK() {
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

	for _, j := range jj {
		if err := opts.JobQueue.AddJob(j); err != nil {
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
