package jobserver

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/interline-io/transitland-server/internal/util"
	"github.com/interline-io/transitland-server/server/jobs"
	"github.com/interline-io/transitland-server/server/jobs/artifactjob"
	"github.com/interline-io/transitland-server/server/model"
)

// NewServer creates a simple api for submitting and running jobs.
func NewServer(queueName string, workers int) (http.Handler, error) {
	r := chi.NewRouter()

	// Legacy job endpoints (backward compatibility)
	r.HandleFunc("/add", addJobRequest)
	r.HandleFunc("/run", runJobRequest)

	// New artifact job endpoints
	r.HandleFunc("/artifact/submit", addArtifactJobRequest)
	r.HandleFunc("/artifact/status/{jobID}", getArtifactJobStatus)
	r.HandleFunc("/artifact/job_registry", listArtifactJobs)

	return r, nil
}

// job response
type jobResponse struct {
	Status  string   `json:"status"`
	Success bool     `json:"success"`
	Error   string   `json:"error,omitempty"`
	Job     jobs.Job `json:"job"`
}

// artifact job request
type artifactJobRequest struct {
	JobName string                 `json:"job_name"`
	Args    map[string]interface{} `json:"args"`
	Env     map[string]string      `json:"env,omitempty"`
}

// artifact job response
type artifactJobResponse struct {
	Status  string                 `json:"status"`
	Success bool                   `json:"success"`
	Error   string                 `json:"error,omitempty"`
	JobID   interface{}            `json:"job_id,omitempty"`
	Job     *artifactjob.JobResult `json:"job,omitempty"`
}

// addJobRequest adds the request to the appropriate queue
func addJobRequest(w http.ResponseWriter, req *http.Request) {
	ctx := req.Context()
	job, err := requestGetJob(req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// add job to queue
	ret := jobResponse{
		Job: job,
	}
	if jobQueue := model.ForContext(ctx).JobQueue; jobQueue == nil {
		ret.Status = "failed"
		ret.Error = "no job queue available"
	} else if err := jobQueue.AddJob(ctx, job); err != nil {
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
	ctx := req.Context()
	job, err := requestGetJob(req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// run job directly
	ret := jobResponse{
		Job: job,
	}
	if jobQueue := model.ForContext(ctx).JobQueue; jobQueue == nil {
		ret.Status = "failed"
		ret.Error = "no job queue available"
	} else if err := jobQueue.RunJob(ctx, job); err != nil {
		ret.Status = "failed"
		ret.Error = err.Error()
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

// addArtifactJobRequest submits an artifact job to the registry
func addArtifactJobRequest(w http.ResponseWriter, req *http.Request) {
	ctx := req.Context()

	// Parse artifact job request
	var artifactReq artifactJobRequest
	if err := json.NewDecoder(req.Body).Decode(&artifactReq); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	if artifactReq.JobName == "" {
		http.Error(w, "job_name is required", http.StatusBadRequest)
		return
	}

	// Get artifact job registry from context
	registry := model.ForContext(ctx).ArtifactJobRegistry
	if registry == nil {
		http.Error(w, "no artifact job registry available", http.StatusInternalServerError)
		return
	}

	// Submit the job
	result, err := registry.SubmitJob(ctx, artifactReq.JobName, artifactReq.Args)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Return success response
	response := artifactJobResponse{
		Status:  "submitted",
		Success: true,
		JobID:   result,
	}

	if rj, err := json.Marshal(response); err != nil {
		util.WriteJsonError(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	} else {
		w.Header().Set("Content-Type", "application/json")
		w.Write(rj)
	}
}

// getArtifactJobStatus retrieves the status of an artifact job
func getArtifactJobStatus(w http.ResponseWriter, req *http.Request) {
	ctx := req.Context()

	// Extract job ID from URL
	jobID := chi.URLParam(req, "jobID")
	if jobID == "" {
		http.Error(w, "job ID is required", http.StatusBadRequest)
		return
	}

	// Get artifact job registry from context
	registry := model.ForContext(ctx).ArtifactJobRegistry
	if registry == nil {
		http.Error(w, "no artifact job registry available", http.StatusInternalServerError)
		return
	}

	// Get job status
	jobResult, err := registry.GetJobStatus(ctx, jobID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Return job status
	response := artifactJobResponse{
		Status:  "success",
		Success: true,
		Job:     jobResult,
	}

	if rj, err := json.Marshal(response); err != nil {
		util.WriteJsonError(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	} else {
		w.Header().Set("Content-Type", "application/json")
		w.Write(rj)
	}
}

// listArtifactJobs lists all available artifact jobs
func listArtifactJobs(w http.ResponseWriter, req *http.Request) {
	ctx := req.Context()

	// Get artifact job registry from context
	registry := model.ForContext(ctx).ArtifactJobRegistry
	if registry == nil {
		http.Error(w, "no artifact job registry available", http.StatusInternalServerError)
		return
	}

	// List available jobs
	availableJobs := registry.ListJobs()

	// Return job list
	response := map[string]interface{}{
		"status": "success",
		"jobs":   availableJobs,
	}

	if rj, err := json.Marshal(response); err != nil {
		util.WriteJsonError(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	} else {
		w.Header().Set("Content-Type", "application/json")
		w.Write(rj)
	}
}
