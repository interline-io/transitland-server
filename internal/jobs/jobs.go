package jobs

import (
	"context"

	"github.com/interline-io/transitland-server/model"
)

// Job queue
type JobQueue interface {
	AddJob(Job) error
	AddWorker(GetWorker, JobOptions, int) error
	Run() error
	Stop() error
}

// Job defines a single job
type Job struct {
	JobType string     `json:"job_type"`
	Feed    string     `json:"feed"`
	URL     string     `json:"url"`
	Args    []string   `json:"args"`
	Opts    JobOptions `json:"-"`
}

// JobOptions is configuration passed to worker.
type JobOptions struct {
	Finder   model.Finder
	RTFinder model.RTFinder
	JobQueue JobQueue
}

// GetWorker returns a new worker for this job type
type GetWorker func(Job) (JobWorker, error)

// JobWorker defines a job worker
type JobWorker interface {
	Run(context.Context, Job) error
}
