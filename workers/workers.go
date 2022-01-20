package workers

import (
	"context"
	"errors"

	"github.com/go-redis/redis/v8"
	"github.com/interline-io/transitland-server/config"
	"github.com/interline-io/transitland-server/model"
	"github.com/interline-io/transitland-server/rtcache"
)

type Job = config.Job

// JobWorker defines a job worker
type JobWorker interface {
	Run(context.Context, JobOptions, rtcache.Job) error
}

// JobOptions is configuration passed to worker.
type JobOptions struct {
	finder   model.Finder
	rtfinder model.RTFinder
	jobs     config.JobQueue
}

// JobRunner works with JobQueue to run jobs with local configuration options.
type JobRunner struct {
	QueueName string
	Workers   int
	cfg       config.Config
	jq        config.JobQueue
	finder    model.Finder
	rtfinder  model.RTFinder
}

// NewJobRunner returns a new configured JobRunner listening to the specified queue.
func NewJobRunner(cfg config.Config, finder model.Finder, rtfinder model.RTFinder, jq config.JobQueue, queueName string, workers int) (*JobRunner, error) {
	j := JobRunner{
		QueueName: queueName,
		Workers:   workers,
		cfg:       cfg,
		jq:        jq,
		finder:    finder,
		rtfinder:  rtfinder,
	}
	return &j, nil
}

// AddJob adds a job to the queue.
func (j *JobRunner) AddJob(job rtcache.Job) error {
	return j.jq.AddJob(job)
}

// RunJob runs a job immediately.
func (j *JobRunner) RunJob(job rtcache.Job) error {
	r, err := GetWorker(job)
	if err != nil {
		return err
	}
	ctx := context.TODO()
	args := JobOptions{
		finder: j.finder,
		jobs:   j.jq,
	}
	return r.Run(ctx, args, job)
}

func (j *JobRunner) RunWorkers() error {
	// Create a new instance each time this is called.
	redisClient := redis.NewClient(&redis.Options{Addr: j.cfg.RT.RedisURL})
	rtJobs := NewRedisJobs(redisClient, j.QueueName)
	rtJobs.AddWorker(j.RunJob, 1)
	return rtJobs.Run()
}

// GetWorker returns the correct worker type for this job.
func GetWorker(job rtcache.Job) (JobWorker, error) {
	var r JobWorker
	class := job.JobType
	switch class {
	case "rt-enqueue":
		r = &RTEnqueueWorker{}
	case "rt-fetch":
		r = &RTFetchWorker{}
	default:
		return nil, errors.New("unknown job type")
	}
	return r, nil
}
