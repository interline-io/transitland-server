package jobs

import (
	"context"
	"os"
	"strconv"
	"time"

	workers "github.com/digitalocean/go-workers2"
	"github.com/go-redis/redis/v8"
	"github.com/interline-io/transitland-lib/log"
)

// RedisJobs is a simple wrapper around go-workers
type RedisJobs struct {
	queueName   string
	producer    *workers.Producer
	manager     *workers.Manager
	client      *redis.Client
	middlewares []JobMiddleware
}

func NewRedisJobs(client *redis.Client, queueName string) *RedisJobs {
	f := RedisJobs{
		queueName: queueName,
		client:    client,
	}
	f.AddMiddleware(newLog())
	return &f
}

func (f *RedisJobs) AddMiddleware(mwf JobMiddleware) {
	f.middlewares = append(f.middlewares, mwf)
}

func (f *RedisJobs) AddJob(job Job) error {
	if f.producer == nil {
		var err error
		f.producer, err = workers.NewProducerWithRedisClient(workers.Options{
			ProcessID: strconv.Itoa(os.Getpid()),
		}, f.client)
		if err != nil {
			return err
		}
	}
	_, err := f.producer.Enqueue(f.queueName, job.JobType, job.JobArgs)
	return err
}

func (f *RedisJobs) getManager() (*workers.Manager, error) {
	var err error
	if f.manager == nil {
		f.manager, err = workers.NewManagerWithRedisClient(workers.Options{
			ProcessID: strconv.Itoa(os.Getpid()),
		}, f.client)
	}
	return f.manager, err
}

func (f *RedisJobs) AddWorker(getWorker GetWorker, jo JobOptions, count int) error {
	manager, err := f.getManager()
	if err != nil {
		return err
	}
	processMessage := func(msg *workers.Msg) error {
		jargs, err := msg.Args().Map()
		if err != nil {
			return err
		}
		job := Job{JobType: msg.Class(), JobArgs: jargs}
		w, err := getWorker(job)
		if err != nil {
			return err
		}
		t1 := time.Now()
		job.Opts = jo
		job.Opts.Logger = log.Logger.With().Str("job_type", job.JobType).Str("job_id", msg.Jid()).Logger()
		job.Opts.Logger.Info().Msg("job: started")
		if err := w.Run(context.TODO(), job); err != nil {
			job.Opts.Logger.Error().Err(err).Msg("job: error")
			return err
		}
		job.Opts.Logger.Info().Int64("job_time_ms", (time.Now().UnixNano()-t1.UnixNano())/1e6).Msg("job: completed")
		return nil
	}
	manager.AddWorker(f.queueName, count, processMessage)
	return nil
}

func (f *RedisJobs) Run() error {
	log.Infof("jobs: running")
	manager, err := f.getManager()
	if err == nil {
		// Blocks
		manager.Run()
	}
	return err
}

func (f *RedisJobs) Stop() error {
	log.Infof("jobs: stopping")
	manager, err := f.getManager()
	if err == nil {
		manager.Stop()
	}
	return err
}
