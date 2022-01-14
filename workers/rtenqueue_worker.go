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
	for _, ent := range qents {
		fmt.Println("found:", ent)
		job := rtcache.Job{
			JobType: "rt-fetch",
			Args:    []string{ent.FeedID, ent.URLs.RealtimeTripUpdates},
		}
		fmt.Println("...enqueue:", job)
		if err := opts.jobs.AddJob(job); err != nil {
			return err
		}
	}
	return nil
}
