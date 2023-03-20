package jobs

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"sync/atomic"

	"github.com/interline-io/transitland-lib/log"
)

var jobCounter = uint64(0)

type LocalJobs struct {
	jobs        chan Job
	jobfuncs    []func(Job) error
	running     bool
	middlewares []JobMiddleware
}

func NewLocalJobs() *LocalJobs {
	f := &LocalJobs{
		jobs: make(chan Job, 1000),
	}
	f.middlewares = append(f.middlewares, newLog())
	return f
}

func (f *LocalJobs) AddMiddleware(mwf JobMiddleware) {
	f.middlewares = append(f.middlewares, mwf)
}

func (f *LocalJobs) AddJob(job Job) error {
	if f.jobs == nil {
		return errors.New("closed")
	}
	f.jobs <- job
	log.Info().Interface("job", job).Msg("jobs: added job") // ("jobs: added job '%s'\n", job.Feed)
	return nil
}

func (f *LocalJobs) AddWorker(getWorker GetWorker, jo JobOptions, count int) error {
	if f.running {
		return errors.New("already running")
	}
	processMessage := func(job Job) error {
		job = Job{JobType: job.JobType, JobArgs: job.JobArgs, Opts: jo, jobId: fmt.Sprintf("%d", atomic.AddUint64(&jobCounter, 1))}
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
		return w.Run(context.TODO(), job)
	}
	log.Infof("jobs: created job listener")
	for i := 0; i < count; i++ {
		f.jobfuncs = append(f.jobfuncs, processMessage)
	}
	return nil
}

func (f *LocalJobs) Run() error {
	if f.running {
		return errors.New("already running")
	}
	log.Infof("jobs: running")
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
	log.Infof("jobs: stopping")
	close(f.jobs)
	f.running = false
	f.jobs = nil
	return nil
}
