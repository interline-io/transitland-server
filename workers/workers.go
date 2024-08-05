package workers

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/interline-io/transitland-mw/jobs"
	"github.com/interline-io/transitland-server/internal/util"
	"github.com/interline-io/transitland-server/model"
)

// GetWorker returns the correct worker type for this job.
func GetWorker(job jobs.Job) (jobs.JobWorker, error) {
	var r jobs.JobWorker
	class := job.JobType
	switch class {
	case "fetch-enqueue":
		r = &FetchEnqueueWorker{}
	case "rt-fetch":
		r = &RTFetchWorker{}
	case "static-fetch":
		r = &StaticFetchWorker{}
	case "gbfs-fetch":
		r = &GbfsFetchWorker{}
	case "test-ok":
		r = &testOkWorker{}
	case "test-fail":
		r = &testFailWorker{}
	default:
		return nil, errors.New("unknown job type")
	}
	// Load json
	jw, err := json.Marshal(job.JobArgs)
	if err != nil {
		return nil, err
	}
	if err := json.Unmarshal(jw, r); err != nil {
		return nil, err
	}
	return r, nil
}

// NewServer creates a simple api for submitting and running jobs.
func NewServer(queueName string, workers int) (http.Handler, error) {
	r := chi.NewRouter()
	r.HandleFunc("/add", addJobRequest)
	r.HandleFunc("/run", runJobRequest)
	return r, nil
}

// job response
type jobResponse struct {
	Status  string   `json:"status"`
	Success bool     `json:"success"`
	Error   string   `json:"error,omitempty"`
	Job     jobs.Job `json:"job"`
}

// addJobRequest adds the request to the appropriate queue
func addJobRequest(w http.ResponseWriter, req *http.Request) {
	job, err := requestGetJob(req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// add job to queue
	ret := jobResponse{
		Job: job,
	}
	if jobQueue := model.ForContext(req.Context()).JobQueue; jobQueue == nil {
		ret.Status = "failed"
		ret.Error = "no job queue available"
	} else if err := jobQueue.AddJob(job); err != nil {
		ret.Status = "failed"
		ret.Error = err.Error()
	} else {
		ret.Status = "added"
		ret.Success = true
	}
	writeJobResponse(ret, w)
}

// runJobRequest runs the job directly
func runJobRequest(w http.ResponseWriter, req *http.Request) {
	job, err := requestGetJob(req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// run job directly
	ret := jobResponse{
		Job: job,
	}
	wk, err := GetWorker(job)
	if err != nil {
		ret.Error = err.Error()
		ret.Status = "failed"
		ret.Success = false
	} else if err := wk.Run(req.Context(), job); err != nil {
		ret.Error = err.Error()
		ret.Status = "failed"
		ret.Success = false
	} else {
		ret.Status = "completed"
		ret.Success = true
	}
	writeJobResponse(ret, w)
}

// requestGetJob parses job from request body
func requestGetJob(req *http.Request) (jobs.Job, error) {
	var job jobs.Job
	err := json.NewDecoder(req.Body).Decode(&job)
	if err != nil {
		return job, errors.New("error parsing body")
	}
	// check worker type
	if _, err := GetWorker(job); err != nil {
		return job, err
	}
	return job, nil
}

// writeJobResponse writes job response
func writeJobResponse(ret jobResponse, w http.ResponseWriter) {
	if rj, err := json.Marshal(ret); err != nil {
		util.WriteJsonError(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	} else {
		w.Write(rj)
	}
}

///////////

type cfgMiddleware struct {
	jobs.JobWorker
	cfg model.Config
}

func (w *cfgMiddleware) Run(ctx context.Context, job jobs.Job) error {
	return w.JobWorker.Run(model.WithConfig(ctx, w.cfg), job)
}

func newCfgMiddleware(cfg model.Config) jobs.JobMiddleware {
	return func(w jobs.JobWorker) jobs.JobWorker {
		return &cfgMiddleware{cfg: cfg, JobWorker: w}
	}
}
