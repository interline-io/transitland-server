package workers

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/interline-io/transitland-server/config"
	"github.com/interline-io/transitland-server/internal/jobs"
	"github.com/interline-io/transitland-server/model"
)

// GetWorker returns the correct worker type for this job.
func GetWorker(job jobs.Job) (jobs.JobWorker, error) {
	var r jobs.JobWorker
	class := job.JobType
	switch class {
	case "rt-enqueue":
		r = &RTEnqueueWorker{}
	case "rt-fetch":
		r = &RTFetchWorker{}
	case "test":
		r = &testWorker{}
	default:
		return nil, errors.New("unknown job type")
	}
	return r, nil
}

// NewServer creates a simple api for submitting and running jobs.
func NewServer(cfg config.Config, finder model.Finder, rtFinder model.RTFinder, jq jobs.JobQueue, queueName string, workers int) (http.Handler, error) {
	r := mux.NewRouter()
	jo := jobs.JobOptions{
		JobQueue: jq,
		Finder:   finder,
		RTFinder: rtFinder,
	}
	r.HandleFunc("/add", wrapHandler(addJobRequest, jo, jq))
	r.HandleFunc("/run", wrapHandler(runJobRequest, jo, jq))
	return r, nil
}

func wrapHandler(next func(http.ResponseWriter, *http.Request, jobs.JobOptions, jobs.JobQueue), jo jobs.JobOptions, jq jobs.JobQueue) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		next(w, r, jo, jq)
	})
}

// job response
type jobResponse struct {
	Status  string   `json:"status"`
	Success bool     `json:"success"`
	Error   string   `json:"error,omitempty"`
	Job     jobs.Job `json:"job"`
}

// addJobRequest adds the request to the appropriate queue
func addJobRequest(w http.ResponseWriter, req *http.Request, jo jobs.JobOptions, jq jobs.JobQueue) {
	job, err := requestGetJob(req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	ret := jobResponse{
		Job: job,
	}
	// add job to queue
	if err := jq.AddJob(job); err != nil {
		ret.Error = err.Error()
		ret.Status = "failed"
		ret.Success = false
	} else {
		ret.Status = "added"
		ret.Success = true
	}
	writeJobResponse(ret, w)
}

// runJobRequest runs the job directly
func runJobRequest(w http.ResponseWriter, req *http.Request, jo jobs.JobOptions, jq jobs.JobQueue) {
	job, err := requestGetJob(req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	ret := jobResponse{
		Job: job,
	}
	// run job directly
	wk, err := GetWorker(job)
	if err != nil {
		// failed
		ret.Error = err.Error()
		ret.Status = "failed"
		ret.Success = false
	}
	job.Opts = jo
	if err := wk.Run(context.TODO(), job); err != nil {
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
		http.Error(w, "request failed", http.StatusBadRequest)
		return
	} else {
		w.Write(rj)
	}
}
