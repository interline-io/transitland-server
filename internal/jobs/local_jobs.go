package jobs

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"sync/atomic"
	"time"

	"github.com/interline-io/log"
)

var jobCounter = uint64(0)

type LocalJobs struct {
	jobs           chan Job
	jobfuncs       []func(Job) error
	running        bool
	middlewares    []JobMiddleware
	uniqueJobs     map[string]bool
	uniqueJobsLock sync.Mutex
}

func NewLocalJobs() *LocalJobs {
	f := &LocalJobs{
		jobs:       make(chan Job, 1000),
		uniqueJobs: map[string]bool{},
	}
	f.middlewares = append(f.middlewares, newLog())
	return f
}

func (f *LocalJobs) Use(mwf JobMiddleware) {
	f.middlewares = append(f.middlewares, mwf)
}

func (f *LocalJobs) AddJob(job Job) error {
	if f.jobs == nil {
		return errors.New("closed")
	}
	if job.Unique {
		f.uniqueJobsLock.Lock()
		defer f.uniqueJobsLock.Unlock()
		key, err := job.HexKey()
		if err != nil {
			return err
		}
		if _, ok := f.uniqueJobs[key]; ok {
			log.Trace().Interface("job", job).Msgf("already locked: %s", key)
			return nil
		} else {
			f.uniqueJobs[key] = true
			log.Trace().Interface("job", job).Msgf("locked: %s", key)
		}
	}
	f.jobs <- job
	log.Info().Interface("job", job).Msg("jobs: added job")
	return nil
}

func (f *LocalJobs) processMessage(getWorker GetWorker, jo JobOptions, job Job) error {
	job = Job{
		JobType:     job.JobType,
		JobArgs:     job.JobArgs,
		JobDeadline: job.JobDeadline,
		Unique:      job.Unique,
		Opts:        jo,
		jobId:       fmt.Sprintf("%d", atomic.AddUint64(&jobCounter, 1)),
	}
	now := time.Now().In(time.UTC).Unix()
	if job.JobDeadline > 0 && job.JobDeadline < now {
		log.Trace().Int64("job_deadline", job.JobDeadline).Int64("now", now).Msg("job skipped - deadline in past")
		return nil
	}
	if job.Unique {
		f.uniqueJobsLock.Lock()
		defer f.uniqueJobsLock.Unlock()
		key, err := job.HexKey()
		if err != nil {
			return err
		}
		delete(f.uniqueJobs, key)
		log.Trace().Interface("job", job).Msgf("unlocked: %s", key)
	}
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

func (f *LocalJobs) AddWorker(queue string, getWorker GetWorker, jo JobOptions, count int) error {
	if f.running {
		return errors.New("already running")
	}
	processMessage := func(job Job) error {
		return f.processMessage(getWorker, jo, job)
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
				if err := jf(job); err != nil {
					log.Trace().Err(err).Msg("job failed")
				}
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
