package artifactjob

import "context"

// Store defines the interface for persisting artifacts
type Store interface {
	// Artifact operations
	SaveArtifacts(ctx context.Context, jobID int64, artifacts []*Artifact) error
	GetArtifact(ctx context.Context, id string) (*Artifact, error)
	ListArtifacts(ctx context.Context, jobID int64) ([]*Artifact, error)
}
