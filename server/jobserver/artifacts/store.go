package artifacts

import (
	"context"
	"database/sql"
	"fmt"

	sq "github.com/Masterminds/squirrel"
	"github.com/interline-io/transitland-server/server/dbutil"
	"github.com/interline-io/transitland-server/server/jobs/artifactjob"
	"github.com/interline-io/transitland-server/server/jobserver/artifacts/storage"
	"github.com/jmoiron/sqlx"
)

// Store manages artifact storage
type Store struct {
	db      *sqlx.DB
	storage storage.Adapter
}

// NewStore creates a new artifact store
func NewStore(db *sqlx.DB, storage storage.Adapter) *Store {
	return &Store{
		db:      db,
		storage: storage,
	}
}

// SaveArtifacts saves artifacts for a job
func (s *Store) SaveArtifacts(ctx context.Context, jobID int64, artifacts []*artifactjob.Artifact) error {
	tx, err := s.db.BeginTxx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	for _, artifact := range artifacts {
		// For now, we only support inline storage
		// Cloud storage can be added later by implementing the storage.Adapter interface
		if artifact.Type != artifactjob.ArtifactTypeInline {
			return fmt.Errorf("unsupported artifact type: %s", artifact.Type)
		}

		// Ensure inline data is provided
		if artifact.InlineData == nil {
			return fmt.Errorf("no data provided for inline artifact")
		}

		// Use Squirrel for insert
		q := sq.Insert("tl_job_artifacts").
			Columns("id", "job_id", "name", "type", "content_type", "size", "created_at", "inline_data", "cloud_storage_ref").
			Values(artifact.ID, artifact.JobID, artifact.Name, artifact.Type,
				artifact.ContentType, artifact.Size, artifact.CreatedAt,
				artifact.InlineData, artifact.CloudStorageRef)

		// Use dbutil.Insert for consistent query execution
		_, err = dbutil.Insert(ctx, tx, q)
		if err != nil {
			return err
		}
	}

	return tx.Commit()
}

// SaveLogs saves log entries for a job
func (s *Store) SaveLogs(ctx context.Context, jobID int64, logs []*artifactjob.JobLog) error {
	if len(logs) == 0 {
		return nil
	}

	tx, err := s.db.BeginTxx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	for _, log := range logs {
		// Use Squirrel for insert
		q := sq.Insert("tl_job_logs").
			Columns("id", "job_id", "level", "message", "timestamp", "metadata").
			Values(log.ID, log.JobID, log.Level, log.Message, log.Timestamp, log.Metadata)

		// Use dbutil.Insert for consistent query execution
		_, err = dbutil.Insert(ctx, tx, q)
		if err != nil {
			return err
		}
	}

	return tx.Commit()
}

// GetLogs retrieves logs for a job
func (s *Store) GetLogs(ctx context.Context, jobID int64) ([]*artifactjob.JobLog, error) {
	// Use Squirrel for select
	q := sq.Select("id", "job_id", "level", "message", "timestamp", "metadata").
		From("tl_job_logs").
		Where(sq.Eq{"job_id": jobID}).
		OrderBy("timestamp ASC")

	// Use dbutil.Select for consistent query execution
	var logs []*artifactjob.JobLog
	if err := dbutil.Select(ctx, s.db, q, &logs); err != nil {
		return nil, err
	}

	return logs, nil
}

// GetArtifact retrieves an artifact by ID
func (s *Store) GetArtifact(ctx context.Context, id string) (*artifactjob.Artifact, error) {
	// Use Squirrel for select
	q := sq.Select("id", "job_id", "name", "type", "content_type", "size", "created_at", "inline_data", "cloud_storage_ref").
		From("tl_job_artifacts").
		Where(sq.Eq{"id": id})

	// Use dbutil.Get for consistent query execution
	artifact := &artifactjob.Artifact{}
	if err := dbutil.Get(ctx, s.db, q, artifact); err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("artifact not found: %s", id)
		}
		return nil, err
	}

	// For cloud storage, get pre-signed URL
	if artifact.CloudStorageRef != nil && s.storage != nil {
		// Use the path as the key for storage lookup
		key := artifact.CloudStorageRef.Path
		if key == "" {
			key = artifact.Name // Fallback to artifact name
		}

		url, err := s.storage.GetURL(ctx, key)
		if err != nil {
			return nil, fmt.Errorf("failed to get storage URL: %w", err)
		}
		artifact.CloudStorageRef.URL = url
	}

	return artifact, nil
}

// ListArtifacts lists artifacts for a job
func (s *Store) ListArtifacts(ctx context.Context, jobID int64) ([]*artifactjob.Artifact, error) {
	// Use Squirrel for select
	q := sq.Select("id", "job_id", "name", "type", "content_type", "size", "created_at", "inline_data", "cloud_storage_ref").
		From("tl_job_artifacts").
		Where(sq.Eq{"job_id": jobID}).
		OrderBy("created_at ASC")

	// Use dbutil.Select for consistent query execution
	var artifacts []*artifactjob.Artifact
	if err := dbutil.Select(ctx, s.db, q, &artifacts); err != nil {
		return nil, err
	}

	// For cloud storage, get pre-signed URLs
	if s.storage != nil {
		for _, artifact := range artifacts {
			if artifact.CloudStorageRef != nil {
				// Use the path as the key for storage lookup
				key := artifact.CloudStorageRef.Path
				if key == "" {
					key = artifact.Name // Fallback to artifact name
				}

				url, err := s.storage.GetURL(ctx, key)
				if err != nil {
					return nil, fmt.Errorf("failed to get storage URL: %w", err)
				}
				artifact.CloudStorageRef.URL = url
			}
		}
	}

	return artifacts, nil
}
