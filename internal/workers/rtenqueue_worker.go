package workers

import (
	"context"

	"github.com/interline-io/transitland-server/internal/jobs"
	"github.com/interline-io/transitland-server/model"
)

type RTEnqueueWorker struct{}

func (w *RTEnqueueWorker) Run(ctx context.Context, job jobs.Job) error {
	// fmt.Println("enqueue worker!")
	opts := job.Opts
	qents, err := opts.Finder.FindFeeds(nil, nil, nil, &model.FeedFilter{Spec: []string{"gtfs-rt"}})
	if err != nil {
		return err
	}
	var jj []jobs.Job
	for _, ent := range qents {
		for _, target := range ent.AssociatedFeeds {
			if ent.URLs.RealtimeAlerts != "" {
				jj = append(jj, jobs.Job{JobType: "rt-fetch", Args: []string{target, "alerts", ent.URLs.RealtimeAlerts}})
			}
			if ent.URLs.RealtimeTripUpdates != "" {
				jj = append(jj, jobs.Job{JobType: "rt-fetch", Args: []string{target, "trip_updates", ent.URLs.RealtimeTripUpdates}})
			}
			if ent.URLs.RealtimeVehiclePositions != "" {
				jj = append(jj, jobs.Job{JobType: "rt-fetch", Args: []string{target, "vehicle_positions", ent.URLs.RealtimeVehiclePositions}})
			}
		}
	}
	for _, job := range jj {
		if err := opts.JobQueue.AddJob(job); err != nil {
			return err
		}
	}
	return nil
}
