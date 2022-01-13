package rtcache

import (
	"errors"
	"fmt"
	"os"
	"strconv"

	workers "github.com/digitalocean/go-workers2"
	"github.com/go-redis/redis/v8"
)

// RedisJobs is a simple wrapper around go-workers
type RedisJobs struct {
	redisUrl      string
	jobQueueTopic string
	jobs          chan Job
	manager       *workers.Manager
	client        *redis.Client
}

func NewRedisJobs(client *redis.Client) *RedisJobs {
	f := RedisJobs{
		client:        client,
		jobQueueTopic: "default",
		jobs:          make(chan Job),
	}
	return &f
}

func (f *RedisJobs) AddJob(job Job) error {
	manager, err := f.start()
	if err != nil {
		return err
	}
	producer := manager.Producer()
	producer.Enqueue(f.jobQueueTopic, job.JobType, []string{job.Feed, job.URL})
	fmt.Printf("jobs '%s': added job '%s'\n", f.jobQueueTopic, job.Feed)
	return nil
}

func (f *RedisJobs) Listen() (chan Job, error) {
	return f.jobs, nil
}

func (f *RedisJobs) Close() error {
	if f.manager != nil {
		f.manager.Stop()
	}
	close(f.jobs)
	return nil
}

func (f *RedisJobs) start() (*workers.Manager, error) {
	if f.manager != nil {
		return f.manager, nil
	}
	manager, err := workers.NewManagerWithRedisClient(workers.Options{
		ServerAddr: f.redisUrl,
		ProcessID:  strconv.Itoa(os.Getpid()),
	}, f.client)
	if err != nil {
		return nil, err
	}
	processMessage := func(msg *workers.Msg) error {
		jargs := msg.Args().MustStringArray()
		if len(jargs) != 2 {
			return errors.New("incorrect args")
		}
		job := Job{JobType: msg.Class(), Feed: jargs[0], URL: jargs[1]}
		f.jobs <- job
		return nil
	}
	manager.AddWorker(f.jobQueueTopic, 1, processMessage)
	f.manager = manager
	go manager.Run()
	return manager, err
}
