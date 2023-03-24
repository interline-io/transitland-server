package metrics

import (
	"context"

	"github.com/interline-io/transitland-server/internal/jobs"
)

func NewJobMiddleware(queue string, m JobMetric) jobs.JobMiddleware {
	return func(w jobs.JobWorker) jobs.JobWorker {
		return &jobWorker{
			jobWorker: w,
			jobMetric: m,
		}
	}
}

type jobWorker struct {
	jobMetric JobMetric
	jobWorker jobs.JobWorker
}

func (w *jobWorker) Run(ctx context.Context, job jobs.Job) error {
	w.jobMetric.AddStartedJob(job.JobType)
	err := w.jobWorker.Run(ctx, job)
	if err != nil {
		w.jobMetric.AddCompletedJob(job.JobType, false)
	} else {
		w.jobMetric.AddCompletedJob(job.JobType, true)
	}
	return err
}
