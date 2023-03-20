package metrics

import (
	"context"

	"github.com/interline-io/transitland-lib/log"
	"github.com/interline-io/transitland-server/internal/jobs"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

type jobMiddleware struct {
	registry   prometheus.Registerer
	jobsTotal  *prometheus.CounterVec
	jobsOk     *prometheus.CounterVec
	jobsFailed *prometheus.CounterVec
	jobs.JobWorker
}

func (w *jobMiddleware) Run(ctx context.Context, job jobs.Job) error {
	w.jobsTotal.With(prometheus.Labels{"class": job.JobType}).Add(1)
	err := w.JobWorker.Run(ctx, job)
	if err != nil {
		w.jobsFailed.With(prometheus.Labels{"class": job.JobType}).Add(1)
	} else {
		w.jobsOk.With(prometheus.Labels{"class": job.JobType}).Add(1)
	}
	return err
}

// New returns a Middleware interface.
func NewJobMiddleware(registry prometheus.Registerer, queue string) jobs.JobMiddleware {
	log.Infof("Registering metrics jobs middleware: %s", queue)
	reg := prometheus.WrapRegistererWith(prometheus.Labels{"queue": queue}, registry)
	jobsTotal := promauto.With(reg).NewCounterVec(
		prometheus.CounterOpts{
			Name: "jobs_processed",
			Help: "Total number of jobs processed",
		}, []string{"class"},
	)
	jobsOk := promauto.With(reg).NewCounterVec(
		prometheus.CounterOpts{
			Name: "jobs_ok",
			Help: "Number of jobs completed successfully",
		}, []string{"class"},
	)

	jobsFailed := promauto.With(reg).NewCounterVec(
		prometheus.CounterOpts{
			Name: "jobs_failed",
			Help: "Failed number of jobs",
		}, []string{"class"},
	)
	return func(w jobs.JobWorker) jobs.JobWorker {
		return &jobMiddleware{
			registry:   registry,
			JobWorker:  w,
			jobsTotal:  jobsTotal,
			jobsOk:     jobsOk,
			jobsFailed: jobsFailed,
		}
	}
}
