package workers

import (
	"context"
	"fmt"

	"github.com/interline-io/transitland-server/model"
)

type RTEnqueueWorker struct{}

func (w *RTEnqueueWorker) Run(ctx context.Context, opts JobOptions, job Job) error {
	fmt.Println("enqueue worker!")
	qents, err := opts.finder.FindFeeds(nil, nil, nil, &model.FeedFilter{Spec: []string{"gtfs-rt"}})
	if err != nil {
		return err
	}
	var jobs []Job
	for _, ent := range qents {
		for _, target := range ent.AssociatedFeeds {
			if ent.URLs.RealtimeAlerts != "" {
				jobs = append(jobs, Job{JobType: "rt-fetch", Args: []string{target, "alerts", ent.URLs.RealtimeAlerts}})
			}
			if ent.URLs.RealtimeTripUpdates != "" {
				jobs = append(jobs, Job{JobType: "rt-fetch", Args: []string{target, "trip_updates", ent.URLs.RealtimeTripUpdates}})
			}
			if ent.URLs.RealtimeVehiclePositions != "" {
				jobs = append(jobs, Job{JobType: "rt-fetch", Args: []string{target, "vehicle_positions", ent.URLs.RealtimeVehiclePositions}})
			}
		}
	}
	for _, job := range jobs {
		if err := opts.jobs.AddJob(job); err != nil {
			return err
		}
	}
	return nil
}
