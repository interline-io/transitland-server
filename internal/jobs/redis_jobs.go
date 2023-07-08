package jobs

import (
	"context"
	"errors"
	"os"
	"strconv"

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
	f.Use(newLog())
	return &f
}

func (f *RedisJobs) Use(mwf JobMiddleware) {
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
		job := Job{JobType: msg.Class(), JobArgs: jargs, Opts: jo, jobId: msg.Jid()}
		w, err := getWorker(job)
		if err != nil {
			return err
		}
		if w == nil {
			return errors.New("no job")
		}
		for _, mwf := range f.middlewares {
			w = mwf(w)
			if w == nil {
				return errors.New("no job")
			}
		}
		if err := w.Run(context.TODO(), job); err != nil {
			log.Trace().Err(err).Msg("job failed")
		}
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
