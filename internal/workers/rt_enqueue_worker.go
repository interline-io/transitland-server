package workers

import (
	"context"

	"github.com/interline-io/transitland-server/internal/jobs"
	"github.com/interline-io/transitland-server/model"
)

type RTEnqueueWorker struct {
	FeedID *string `json:"feed_id"`
}

func (w *RTEnqueueWorker) Run(ctx context.Context, job jobs.Job) error {
	// Get all feeds, filter with RT urls
	opts := job.Opts
	rtfeeds, err := opts.Finder.FindFeeds(ctx, nil, nil, nil, nil, &model.FeedFilter{OnestopID: w.FeedID})
	if err != nil {
		return err
	}
	var jj []jobs.Job
	for _, ent := range rtfeeds {
		fid := ent.FeedID
		target := fid
		if ent.URLs.RealtimeAlerts != "" {
			jj = append(jj, jobs.Job{JobType: "rt-fetch", JobArgs: jobs.JobArgs{"target": target, "source_type": "realtime_alerts", "url": ent.URLs.RealtimeAlerts, "source_feed_id": fid}})
		}
		if ent.URLs.RealtimeTripUpdates != "" {
			jj = append(jj, jobs.Job{JobType: "rt-fetch", JobArgs: jobs.JobArgs{"target": target, "source_type": "realtime_trip_updates", "url": ent.URLs.RealtimeTripUpdates, "source_feed_id": fid}})
		}
		if ent.URLs.RealtimeVehiclePositions != "" {
			jj = append(jj, jobs.Job{JobType: "rt-fetch", JobArgs: jobs.JobArgs{"target": target, "source_type": "realtime_vehicle_positions", "url": ent.URLs.RealtimeVehiclePositions, "source_feed_id": fid}})
		}
	}
	for _, job := range jj {
		if err := opts.JobQueue.AddJob(job); err != nil {
			return err
		}
	}
	return nil
}
