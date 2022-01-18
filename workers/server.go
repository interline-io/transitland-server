package workers

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/interline-io/transitland-server/config"
	"github.com/interline-io/transitland-server/rtcache"
)

// NewServer creates a simple api for submitting and running jobs.
func NewServer(cfg config.Config, queueName string, workers int) (http.Handler, error) {
	r := mux.NewRouter()
	runner, err := NewJobRunner(cfg, queueName, workers)
	if err != nil {
		return nil, err
	}
	fmt.Println("new runner:", runner)
	r.HandleFunc("/add", wrapHandler(addJobRequest, runner))
	r.HandleFunc("/run", wrapHandler(runJobRequest, runner))
	return r, nil
}

func wrapHandler(next func(*JobRunner, http.ResponseWriter, *http.Request), jr *JobRunner) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		next(jr, w, r)
	})
}

// job response
type jobResponse struct {
	Status  string      `json:"status"`
	Success bool        `json:"success"`
	Error   string      `json:"error,omitempty"`
	Job     rtcache.Job `json:"job"`
}

// addJobRequest adds the request to the appropriate queue
func addJobRequest(jr *JobRunner, w http.ResponseWriter, req *http.Request) {
	job, err := requestGetJob(req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	ret := jobResponse{
		Job: job,
	}
	// add job to queue
	if err := jr.AddJob(job); err != nil {
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
func runJobRequest(jr *JobRunner, w http.ResponseWriter, req *http.Request) {
	job, err := requestGetJob(req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	ret := jobResponse{
		Job: job,
	}
	// run job directly
	if err := jr.RunJob(job); err != nil {
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
func requestGetJob(req *http.Request) (rtcache.Job, error) {
	var job rtcache.Job
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
