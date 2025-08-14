package artifacts

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/interline-io/transitland-server/server/jobs/artifactjob"
	"github.com/interline-io/transitland-server/server/jobserver"
)

// Server handles HTTP requests for job artifacts
type Server struct {
	store *Store
}

// NewServer creates a new artifacts server
func NewServer(store *Store) *Server {
	return &Server{
		store: store,
	}
}

// Routes returns the router for artifact endpoints
func (s *Server) Routes() chi.Router {
	r := chi.NewRouter()

	// Get auth provider
	authProvider := jobserver.GetAuthProvider()

	// Add auth middleware
	r.Use(authProvider.Middleware)

	// Add routes
	r.Get("/jobs/{jobID}/artifacts", s.listArtifacts)
	r.Get("/artifacts/{artifactID}", s.getArtifact)
	return r
}

// listArtifacts handles listing artifacts for a job
func (s *Server) listArtifacts(w http.ResponseWriter, r *http.Request) {
	jobID, err := strconv.ParseInt(chi.URLParam(r, "jobID"), 10, 64)
	if err != nil {
		http.Error(w, "Invalid job ID", http.StatusBadRequest)
		return
	}

	// Check access
	authProvider := jobserver.GetAuthProvider()
	if !authProvider.CheckJobAccess(r.Context(), jobID) {
		http.Error(w, "Forbidden", http.StatusForbidden)
		return
	}

	// Get artifacts from store
	artifacts, err := s.store.ListArtifacts(r.Context(), jobID)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to list artifacts: %v", err), http.StatusInternalServerError)
		return
	}

	// Return artifacts
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(artifacts)
}

// getArtifact handles retrieving a single artifact
func (s *Server) getArtifact(w http.ResponseWriter, r *http.Request) {
	artifactID := chi.URLParam(r, "artifactID")

	// Get artifact from store
	artifact, err := s.store.GetArtifact(r.Context(), artifactID)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to get artifact: %v", err), http.StatusNotFound)
		return
	}

	// Check access
	authProvider := jobserver.GetAuthProvider()
	if !authProvider.CheckJobAccess(r.Context(), artifact.JobID) {
		http.Error(w, "Forbidden", http.StatusForbidden)
		return
	}

	// For inline data, return directly
	if artifact.Type == artifactjob.ArtifactTypeInline && artifact.InlineData != nil {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(artifact.InlineData.Data)
		return
	}

	// For cloud storage, redirect to URL if available
	if artifact.CloudStorageRef != nil && artifact.CloudStorageRef.URL != "" {
		http.Redirect(w, r, artifact.CloudStorageRef.URL, http.StatusTemporaryRedirect)
		return
	}

	// If no direct access available, return artifact metadata
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(artifact)
}
