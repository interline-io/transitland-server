package jobs

import (
	"context"
	"os"
	"strconv"

	workers "github.com/digitalocean/go-workers2"
	"github.com/go-redis/redis/v8"
)

// RedisJobs is a simple wrapper around go-workers
type RedisJobs struct {
	queueName string
	producer  *workers.Producer
	manager   *workers.Manager
	client    *redis.Client
}

func NewRedisJobs(client *redis.Client, queueName string) *RedisJobs {
	f := RedisJobs{
		queueName: queueName,
		client:    client,
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
	_, err := f.producer.Enqueue(f.queueName, job.JobType, job.Args)
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
		jargs, _ := msg.Args().StringArray()
		job := Job{JobType: msg.Class(), Args: jargs}
		w, err := getWorker(job)
		if err != nil {
			return err
		}
		job.Opts = jo
		return w.Run(context.TODO(), job)

	}
	manager.AddWorker(f.queueName, count, processMessage)
	return nil
}

func (f *RedisJobs) Run() error {
	// fmt.Println("jobs: running")
	manager, err := f.getManager()
	if err == nil {
		// Blocks
		manager.Run()
	}
	return err
}

func (f *RedisJobs) Stop() error {
	// fmt.Println("jobs: stopping")
	manager, err := f.getManager()
	if err == nil {
		manager.Stop()
	}
	return err
}
