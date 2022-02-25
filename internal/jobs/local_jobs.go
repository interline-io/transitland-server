package jobs

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"sync/atomic"
	"time"

	"github.com/interline-io/transitland-lib/log"
)

var jobCounter = uint64(0)

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
	log.Info().Interface("job", job).Msg("jobs: added job") // ("jobs: added job '%s'\n", job.Feed)
	return nil
}

func (f *LocalJobs) AddWorker(getWorker GetWorker, jo JobOptions, count int) error {
	if f.running {
		return errors.New("already running")
	}
	processMessage := func(job Job) error {
		w, err := getWorker(job)
		if err != nil {
			return err
		}
		t1 := time.Now()
		jobId := atomic.AddUint64(&jobCounter, 1)
		job.Opts = jo
		job.Opts.Logger = log.Logger.With().Str("job_type", job.JobType).Str("job_id", fmt.Sprintf("%d", jobId)).Logger()
		job.Opts.Logger.Info().Msg("job: started")
		if err := w.Run(context.TODO(), job); err != nil {
			job.Opts.Logger.Error().Err(err).Msg("job: error")
			return err
		}
		job.Opts.Logger.Info().Int64("job_time_ms", (time.Now().UnixNano()-t1.UnixNano())/1e6).Msg("job: completed")
		return nil
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
