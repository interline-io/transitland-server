package jobs

import (
	"context"

	"github.com/interline-io/transitland-lib/tl"
	"github.com/interline-io/transitland-server/config"
	"github.com/interline-io/transitland-server/model"
	"github.com/rs/zerolog"
)

type JobArgs map[string]interface{}

// Job queue
type JobQueue interface {
	AddJob(Job) error
	AddWorker(GetWorker, JobOptions, int) error
	Use(JobMiddleware)
	Run() error
	Stop() error
}

// Job defines a single job
type Job struct {
	JobType string     `json:"job_type"`
	JobArgs JobArgs    `json:"job_args"`
	Opts    JobOptions `json:"-"`
	jobId   string
}

// JobOptions is configuration passed to worker.
type JobOptions struct {
	Finder     model.Finder
	RTFinder   model.RTFinder
	GbfsFinder model.GbfsFinder
	JobQueue   JobQueue
	Logger     zerolog.Logger
	Config     config.Config
	Secrets    []tl.Secret
}

// GetWorker returns a new worker for this job type
type GetWorker func(Job) (JobWorker, error)

// JobWorker defines a job worker
type JobWorker interface {
	Run(context.Context, Job) error
}

type JobMiddleware func(JobWorker) JobWorker
