package workers

import (
	"context"

	"github.com/interline-io/transitland-server/internal/jobs"
	"github.com/interline-io/transitland-server/model"
	"github.com/jmoiron/sqlx"
)

type RTEnqueueWorker struct{}

func (w *RTEnqueueWorker) Run(ctx context.Context, job jobs.Job) error {
	opts := job.Opts
	log := job.Opts.Logger
	// Get all operators
	type skey struct {
		RT     string
		Static string
	}
	q := `
	select
		cf2.onestop_id as rt,
		cf1.onestop_id as static
	from current_operators_in_feed coif1
	join current_operators_in_feed coif2 on coif2.resolved_onestop_id = coif1.resolved_onestop_id
	join current_feeds cf1 on cf1.id = coif1.feed_id
	join current_feeds cf2 on cf2.id = coif2.feed_id
	where 
		cf1.spec = 'gtfs'
		and cf2.spec = 'gtfs-rt'
	group by rt,static
	order by rt`
	targets := []skey{}
	if err := sqlx.Select(opts.Finder.DBX(), &targets, q); err != nil {
		return err
	}
	// Get all RT feeds
	rtfeeds, err := opts.Finder.FindFeeds(nil, nil, nil, &model.FeedFilter{Spec: []string{"gtfs-rt"}})
	if err != nil {
		return err
	}
	var jj []jobs.Job
	for _, ent := range rtfeeds {
		fid := ent.FeedID
		if ent.Authorization.Type != "" {
			log.Info().Str("feed_id", fid).Msg("enqueue worker: feed requires auth, skipping")
			continue
		}
		var uniq []string
		for _, sk := range targets {
			if sk.RT == fid {
				uniq = append(uniq, sk.Static)
			}
		}
		log.Info().Str("feed_id", fid).Strs("targets", uniq).Msg("enqueue worker: adding rt-fetch jobs for feed")
		for _, target := range uniq {
			if ent.URLs.RealtimeAlerts != "" {
				jj = append(jj, jobs.Job{JobType: "rt-fetch", Args: []string{target, "alerts", ent.URLs.RealtimeAlerts, fid}})
			}
			if ent.URLs.RealtimeTripUpdates != "" {
				jj = append(jj, jobs.Job{JobType: "rt-fetch", Args: []string{target, "trip_updates", ent.URLs.RealtimeTripUpdates, fid}})
			}
			if ent.URLs.RealtimeVehiclePositions != "" {
				jj = append(jj, jobs.Job{JobType: "rt-fetch", Args: []string{target, "vehicle_positions", ent.URLs.RealtimeVehiclePositions, fid}})
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
