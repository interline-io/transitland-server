package rtcache

import (
	"os"
	"strconv"

	workers "github.com/digitalocean/go-workers2"
	"github.com/go-redis/redis/v8"
)

// RedisJobs is a simple wrapper around go-workers
type RedisJobs struct {
	QueueName string
	jobs      chan Job
	producer  *workers.Producer
	manager   *workers.Manager
	client    *redis.Client
}

func NewRedisJobs(client *redis.Client, queueName string) *RedisJobs {
	f := RedisJobs{
		QueueName: queueName,
		client:    client,
		jobs:      make(chan Job),
	}
	return &f
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
	_, err := f.producer.Enqueue(f.QueueName, job.JobType, job.Args)
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

func (f *RedisJobs) AddWorker(jobfunc func(Job) error, count int) error {
	manager, err := f.getManager()
	if err != nil {
		return err
	}
	processMessage := func(msg *workers.Msg) error {
		jargs, _ := msg.Args().StringArray()
		return jobfunc(Job{JobType: msg.Class(), Args: jargs})
	}
	manager.AddWorker(f.QueueName, 1, processMessage)
	return nil
}

func (f *RedisJobs) Run() error {
	manager, err := f.getManager()
	if err == nil {
		// Blocks
		manager.Run()
	}
	return err
}

func (f *RedisJobs) Stop() error {
	manager, err := f.getManager()
	if err == nil {
		manager.Stop()
	}
	return err
}

func (f *RedisJobs) Listen() (chan Job, error) {
	jobch := make(chan Job)
	return jobch, nil
}

func (f *RedisJobs) Close() error {
	return nil
}
