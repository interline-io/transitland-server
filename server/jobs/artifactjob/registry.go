package artifactjob

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"time"
)

// JobRunner defines the interface for executing an artifact job
type JobRunner interface {
	Run(ctx context.Context, args map[string]any, env map[string]string) error
	GetArtifacts() []Artifact
}

// JobDefinition represents a registered artifact job
type JobDefinition struct {
	Name        string `json:"name"`
	Description string `json:"description"`

	// One of these must be set
	GoRunner     JobRunner `json:"-"`              // For Go-based jobs
	TSScriptPath string    `json:"ts_script_path"` // For TypeScript-based jobs (local file path)

	// Environment variables to pass to the job
	Env map[string]string `json:"env,omitempty"`
}

// JobResult represents the result of an artifact job execution
type JobResult struct {
	JobID       string     `json:"job_id"`
	Status      string     `json:"status"`
	StartedAt   time.Time  `json:"started_at"`
	CompletedAt *time.Time `json:"completed_at,omitempty"`
	Error       string     `json:"error,omitempty"`
	Logs        []string   `json:"logs,omitempty"`
}

// JobArgs represents the arguments for an artifact job in a job queue
type JobArgs struct {
	JobName string                 `json:"job_name"`
	Args    map[string]interface{} `json:"args"`
	Env     map[string]string      `json:"env"`
}

// RiverJobArgs extends JobArgs to implement River's JobArgs interface
type RiverJobArgs struct {
	JobArgs // Embed the base JobArgs
}

// Kind implements River's JobArgs interface
func (r RiverJobArgs) Kind() string {
	return "riverJobArgs"
}

// Registry manages the available artifact jobs
type Registry struct {
	jobs map[string]JobDefinition
}

// ArtifactJobRegistry extends the base Registry with job queue functionality
type ArtifactJobRegistry struct {
	*Registry // Embed the base registry
}

// JobSubmitter defines the interface for submitting jobs to a job queue
type JobSubmitter interface {
	SubmitJob(ctx context.Context, jobName string, args map[string]interface{}) (interface{}, error)
	GetJobStatus(ctx context.Context, jobID string) (*JobResult, error)
	ListJobs() []JobDefinition
}

// NewRegistry creates a new job registry
func NewRegistry() *Registry {
	return &Registry{
		jobs: make(map[string]JobDefinition),
	}
}

// NewArtifactJobRegistry creates a new artifact job registry
func NewArtifactJobRegistry() *ArtifactJobRegistry {
	baseRegistry := NewRegistry()
	return &ArtifactJobRegistry{
		Registry: baseRegistry,
	}
}

// RegisterJob adds a new job to the registry
func (r *Registry) RegisterJob(def JobDefinition) error {
	if def.Name == "" {
		return errors.New("job name is required")
	}
	if def.GoRunner == nil && def.TSScriptPath == "" {
		return errors.New("either GoRunner or TSScriptPath must be provided")
	}
	if def.GoRunner != nil && def.TSScriptPath != "" {
		return errors.New("cannot provide both GoRunner and TSScriptPath")
	}
	r.jobs[def.Name] = def
	return nil
}

// envWrapper wraps a JobRunner to include environment variables
type envWrapper struct {
	runner JobRunner
	env    map[string]string
}

// Run executes the wrapped runner with environment variables
func (w *envWrapper) Run(ctx context.Context, args map[string]any, env map[string]string) error {
	// Merge environment variables (wrapper env takes precedence)
	mergedEnv := make(map[string]string)
	if env != nil {
		for k, v := range env {
			mergedEnv[k] = v
		}
	}
	if w.env != nil {
		for k, v := range w.env {
			mergedEnv[k] = v
		}
	}

	return w.runner.Run(ctx, args, mergedEnv)
}

// GetArtifacts returns artifacts from the wrapped runner
func (w *envWrapper) GetArtifacts() []Artifact {
	return w.runner.GetArtifacts()
}

// GetJob retrieves a job definition by name
func (r *Registry) GetJob(name string) (JobDefinition, error) {
	job, ok := r.jobs[name]
	if !ok {
		return JobDefinition{}, fmt.Errorf("job not found: %s", name)
	}
	return job, nil
}

// ListJobs returns all registered jobs
func (r *Registry) ListJobs() []JobDefinition {
	jobs := make([]JobDefinition, 0, len(r.jobs))
	for _, job := range r.jobs {
		jobs = append(jobs, job)
	}
	return jobs
}

// CreateJobRunner creates a JobRunner from a JobDefinition
func (r *Registry) CreateJobRunner(def JobDefinition) (JobRunner, error) {
	if def.GoRunner != nil {
		return def.GoRunner, nil
	}
	if def.TSScriptPath != "" {
		return NewTSScriptRunner(def.TSScriptPath), nil
	}
	return nil, errors.New("no valid job runner found in definition")
}

// CreateJobRunnerWithEnv creates a JobRunner from a JobDefinition and merges additional environment variables
func (r *Registry) CreateJobRunnerWithEnv(def JobDefinition, additionalEnv map[string]string) (JobRunner, error) {
	runner, err := r.CreateJobRunner(def)
	if err != nil {
		return nil, err
	}

	// Merge environment variables (additionalEnv takes precedence)
	mergedEnv := make(map[string]string)
	if def.Env != nil {
		for k, v := range def.Env {
			mergedEnv[k] = v
		}
	}
	if additionalEnv != nil {
		for k, v := range additionalEnv {
			mergedEnv[k] = v
		}
	}

	// Create a wrapper that includes the merged environment
	return &envWrapper{
		runner: runner,
		env:    mergedEnv,
	}, nil
}

// TSScriptRunner implements JobRunner for TypeScript jobs
type TSScriptRunner struct {
	scriptPath string
	artifacts  []Artifact
}

// NewTSScriptRunner creates a new TypeScript script runner
func NewTSScriptRunner(scriptPath string) *TSScriptRunner {
	return &TSScriptRunner{
		scriptPath: scriptPath,
	}
}

// Run executes the TypeScript script
func (r *TSScriptRunner) Run(ctx context.Context, args map[string]any, env map[string]string) error {
	// Validate script path exists
	if _, err := os.Stat(r.scriptPath); os.IsNotExist(err) {
		return fmt.Errorf("script file not found: %s", r.scriptPath)
	}

	// Write args to temp file
	argsFile, err := os.CreateTemp("", "ts-args-*.json")
	if err != nil {
		return fmt.Errorf("failed to create args file: %w", err)
	}
	defer os.Remove(argsFile.Name())

	argsBytes, err := json.Marshal(args)
	if err != nil {
		return fmt.Errorf("failed to marshal args: %w", err)
	}

	if err := os.WriteFile(argsFile.Name(), argsBytes, 0644); err != nil {
		return fmt.Errorf("failed to write args file: %w", err)
	}

	// Execute script with deno and capture output
	cmd := exec.CommandContext(ctx, "deno", "run", "--allow-read", "--allow-write", "--allow-net", r.scriptPath, argsFile.Name())

	// Set environment variables
	if env != nil {
		for key, value := range env {
			cmd.Env = append(cmd.Env, fmt.Sprintf("%s=%s", key, value))
		}
	}

	// Capture stdout for artifacts
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		// Return stderr as error message
		if stderr.Len() > 0 {
			return fmt.Errorf("script failed: %s", stderr.String())
		}
		return fmt.Errorf("failed to execute TypeScript script: %w", err)
	}

	// Parse stdout as JSON artifacts
	if stdout.Len() > 0 {
		var artifacts []Artifact
		if err := json.Unmarshal(stdout.Bytes(), &artifacts); err != nil {
			return fmt.Errorf("failed to parse artifacts from stdout: %w", err)
		}
		r.artifacts = artifacts
	}

	return nil
}

// GetArtifacts returns artifacts produced by the job
func (r *TSScriptRunner) GetArtifacts() []Artifact {
	return r.artifacts
}
