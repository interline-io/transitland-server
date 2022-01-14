package workers

import (
	"context"
	"errors"

	"github.com/go-redis/redis/v8"
	"github.com/interline-io/transitland-server/config"
	"github.com/interline-io/transitland-server/model"
	"github.com/interline-io/transitland-server/rtcache"
)

type Config struct {
	QueueName string
	Workers   int
	config.Config
}

// configuration to pass to wrapped handler function.
type JobOptions struct {
	jobs  rtcache.JobQueue
	cache rtcache.Cache
	db    model.DBX
}

///////

type JobWorker interface {
	Run(context.Context, JobOptions, rtcache.Job) error
}

type JobRunner struct {
	client *redis.Client
	args   JobOptions
	config Config
}

func NewJobRunner(client *redis.Client, db model.DBX, cfg Config) (*JobRunner, error) {
	j := JobRunner{
		client: client,
		config: cfg,
		args: JobOptions{
			db:    db,
			cache: rtcache.NewRedisCache(client),
			jobs:  rtcache.NewRedisJobs(client, cfg.QueueName),
		}}
	return &j, nil
}

func (j *JobRunner) AddJob(job rtcache.Job) error {
	return j.args.jobs.AddJob(job)
}

func (j *JobRunner) RunJob(job rtcache.Job) error {
	r, err := GetWorker(job)
	if err != nil {
		return err
	}
	ctx := context.TODO()
	return r.Run(ctx, j.args, job)
}

func (j *JobRunner) RunWorkers() error {
	// get a new queue
	rtJobs := rtcache.NewRedisJobs(j.client, j.config.QueueName)
	rtJobs.AddWorker(j.RunJob, j.config.Workers)
	return rtJobs.Run()
}

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
