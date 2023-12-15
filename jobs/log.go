package jobs

import (
	"context"
	"time"

	"github.com/interline-io/log"
)

type jlog struct {
	JobWorker
}

func (w *jlog) Run(ctx context.Context, job Job) error {
	t1 := time.Now()
	job.Opts.Logger = log.Logger.With().Str("job_type", job.JobType).Str("job_id", job.jobId).Logger()
	job.Opts.Logger.Info().Msg("job: started")
	if err := w.JobWorker.Run(ctx, job); err != nil {
		job.Opts.Logger.Error().Err(err).Msg("job: error")
		return err
	}
	job.Opts.Logger.Info().Int64("job_time_ms", (time.Now().UnixNano()-t1.UnixNano())/1e6).Msg("job: completed")
	return nil

}

func newLog() JobMiddleware {
	return func(w JobWorker) JobWorker {
		return &jlog{JobWorker: w}
	}
}
