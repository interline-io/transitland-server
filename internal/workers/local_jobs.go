package workers

import (
	"errors"
	"sync"
)

type LocalJobs struct {
	jobs     chan Job
	jobfuncs []func(Job) error
	running  bool
}

func NewLocalJobs() *LocalJobs {
	return &LocalJobs{
		jobs: make(chan Job, 1000),
	}
}

func (f *LocalJobs) AddJob(job Job) error {
	if f.jobs == nil {
		return errors.New("closed")
	}
	f.jobs <- job
	// fmt.Printf("jobs: added job '%s'\n", job.Feed)
	return nil
}

func (f *LocalJobs) AddWorker(jobfunc func(Job) error, workers int) error {
	if f.running {
		return errors.New("already running")
	}
	// fmt.Printf("jobs: created job listener\n")
	for i := 0; i < workers; i++ {
		f.jobfuncs = append(f.jobfuncs, jobfunc)
	}
	return nil
}

func (f *LocalJobs) Run() error {
	if f.running {
		return errors.New("already running")
	}
	f.running = true
	var wg sync.WaitGroup
	for _, jobfunc := range f.jobfuncs {
		wg.Add(1)
		go func(jf func(Job) error, w *sync.WaitGroup) {
			for job := range f.jobs {
				jf(job)
			}
			wg.Done()
		}(jobfunc, &wg)
	}
	wg.Wait()
	return nil
}

func (f *LocalJobs) Stop() error {
	if !f.running {
		return errors.New("not running")
	}
	// fmt.Println("jobs: stopping")
	close(f.jobs)
	f.running = false
	f.jobs = nil
	return nil
}
