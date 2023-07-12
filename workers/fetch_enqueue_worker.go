package workers

import (
	"context"
	"time"

	"github.com/interline-io/transitland-server/actions"
	"github.com/interline-io/transitland-server/internal/jobs"
	"github.com/interline-io/transitland-server/model"
)

type FetchEnqueueWorker struct{}

func (w *FetchEnqueueWorker) Run(ctx context.Context, job jobs.Job) error {
	opts := job.Opts
	feeds, err := job.Opts.Finder.FindFeeds(ctx, nil, nil, nil, nil, &model.FeedFilter{})
	if err != nil {
		return err
	}
	checkFeed := func(feedId int, fetchWait int64, urlType string, url string) (bool, error) {
		db := opts.Finder.DBX()
		if url == "" {
			return false, nil
		}
		if ok, err := actions.CheckFetchWait(ctx, db, feedId, urlType, time.Duration(fetchWait)*time.Second); err != nil {
			return false, err
		} else if !ok {
			return false, nil
		}
		return true, nil
	}

	var jj []jobs.Job
	jobTime := time.Now()
	_ = jobTime

	// Check static
	for _, feed := range feeds {
		url := feed.URLs.StaticCurrent
		if ok, err := checkFeed(feed.ID, feed.FetchWait.Val, "static_current", url); err != nil {
			return err
		} else if !ok {
			continue
		}
		// TODO: Make this type safe, use StaticFetchWorker{} as job args
		jobArgs := map[string]any{
			"feed_url": url,
			"feed_id":  feed.FeedID,
		}
		jj = append(jj, jobs.Job{JobType: "static-fetch", JobArgs: jobArgs})
	}

	// Check GBFS
	for _, feed := range feeds {
		_ = feed
	}

	// Check RT
	for _, feed := range feeds {
		// Enqueue
		fid := feed.FeedID
		target := fid
		checkUrls := map[string]string{
			"realtime_alerts":            feed.URLs.RealtimeAlerts,
			"realtime_trip_updates":      feed.URLs.RealtimeTripUpdates,
			"realtime_vehicle_positions": feed.URLs.RealtimeTripUpdates,
		}
		for urlType, url := range checkUrls {
			if ok, err := checkFeed(feed.ID, feed.FetchWait.Val, urlType, url); err != nil {
				return err
			} else if !ok {
				continue
			}
			// TODO: Make this type safe, use RTFetchWorker{} as job args
			jobArgs := map[string]any{
				"target":         target,
				"url":            url,
				"source_type":    urlType,
				"source_feed_id": fid,
			}
			jj = append(jj, jobs.Job{JobType: "rt-fetch", JobArgs: jobArgs})
		}
	}
	for _, j := range jj {
		if err := opts.JobQueue.AddJob(j); err != nil {
			return err
		}
	}
	return nil
}
