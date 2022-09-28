package workers

import (
	"context"

	"github.com/interline-io/transitland-server/internal/jobs"
	"github.com/interline-io/transitland-server/model"
)

type GbfsEnqueueWorker struct {
	FeedID *string `json:"feed_id"`
}

func (w *GbfsEnqueueWorker) Run(ctx context.Context, job jobs.Job) error {
	// Get all feeds, filter with RT urls
	opts := job.Opts
	rtfeeds, err := opts.Finder.FindFeeds(ctx, nil, nil, nil, &model.FeedFilter{OnestopID: w.FeedID})
	if err != nil {
		return err
	}
	var jj []jobs.Job
	for _, ent := range rtfeeds {
		fid := ent.FeedID
		target := fid
		if ent.URLs.GbfsAutoDiscovery != "" {
			jj = append(jj, jobs.Job{JobType: "gbfs-fetch", JobArgs: jobs.JobArgs{"target": target, "source_type": "gbfs", "url": ent.URLs.RealtimeAlerts, "source_feed_id": fid}})
		}
	}
	for _, job := range jj {
		if err := opts.JobQueue.AddJob(job); err != nil {
			return err
		}
	}
	return nil
}
