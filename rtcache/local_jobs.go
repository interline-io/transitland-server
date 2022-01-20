package rtcache

import (
	"fmt"
)

type LocalJobs struct {
	jobs chan Job
}

func NewLocalJobs() *LocalJobs {
	return &LocalJobs{
		jobs: make(chan Job, 1000),
	}
}

func (f *LocalJobs) AddJob(job Job) error {
	f.jobs <- job
	fmt.Printf("jobs: added job '%s'\n", job.Feed)
	return nil
}

func (f *LocalJobs) Listen() (chan Job, error) {
	fmt.Printf("jobs: created job listener\n")
	return f.jobs, nil
}

func (f *LocalJobs) Close() error {
	close(f.jobs)
	return nil
}
