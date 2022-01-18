package workers

import (
	"context"
	"fmt"

	sq "github.com/Masterminds/squirrel"
	"github.com/interline-io/transitland-server/find"
	"github.com/interline-io/transitland-server/model"
	"github.com/interline-io/transitland-server/rtcache"
)

type RTEnqueueWorker struct{}

func (w *RTEnqueueWorker) Run(ctx context.Context, opts JobOptions, job rtcache.Job) error {
	fmt.Println("enqueue worker!")
	q := model.Sqrl(opts.db).Select("*").From("current_feeds").Where(sq.Eq{"spec": "gtfs-rt"})
	var qents []*model.Feed
	find.MustSelect(opts.db, q, &qents)
	var jobs []rtcache.Job
	for _, ent := range qents {
		for _, target := range ent.AssociatedFeeds {
			if ent.URLs.RealtimeAlerts != "" {
				jobs = append(jobs, rtcache.Job{JobType: "rt-fetch", Args: []string{target, "alerts", ent.URLs.RealtimeAlerts}})
			}
			if ent.URLs.RealtimeTripUpdates != "" {
				jobs = append(jobs, rtcache.Job{JobType: "rt-fetch", Args: []string{target, "trip_updates", ent.URLs.RealtimeTripUpdates}})
			}
			if ent.URLs.RealtimeVehiclePositions != "" {
				jobs = append(jobs, rtcache.Job{JobType: "rt-fetch", Args: []string{target, "vehicle_positions", ent.URLs.RealtimeVehiclePositions}})
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
