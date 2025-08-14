package artifactjob

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
	"time"

	"github.com/google/uuid"
)

// ArtifactType represents the type of artifact data
type ArtifactType string

const (
	ArtifactTypeInline       ArtifactType = "inline"
	ArtifactTypeCloudStorage ArtifactType = "cloud_storage"
)

// Artifact represents output data produced by a job
type Artifact struct {
	ID          string       `json:"id"`
	JobID       int64        `json:"job_id"` // References river_job.id
	Name        string       `json:"name"`
	Type        ArtifactType `json:"type"`
	ContentType string       `json:"content_type"`
	Size        int64        `json:"size"`
	CreatedAt   time.Time    `json:"created_at"`

	// For inline JSONB data
	InlineData *InlineData `json:"inline_data,omitempty"`

	// For cloud storage references (optional)
	CloudStorageRef *CloudStorageRef `json:"cloud_storage_ref,omitempty"`
}

// InlineData represents inline JSONB data
type InlineData struct {
	Data interface{} `json:"data"`
}

// Value implements the driver.Valuer interface for InlineData
func (i *InlineData) Value() (driver.Value, error) {
	if i == nil {
		return nil, nil
	}
	return json.Marshal(i.Data)
}

// Scan implements the sql.Scanner interface for InlineData
func (i *InlineData) Scan(value interface{}) error {
	if value == nil {
		i.Data = nil
		return nil
	}
	b, ok := value.([]byte)
	if !ok {
		return errors.New("type assertion to []byte failed")
	}
	return json.Unmarshal(b, &i.Data)
}

// CloudStorageRef represents a reference to data in cloud storage
type CloudStorageRef struct {
	// Storage provider identification
	Provider string `json:"provider"` // "azure", "s3", etc.

	// Azure Blob Storage fields
	AccountName string `json:"account_name,omitempty"` // Azure storage account name
	Container   string `json:"container"`              // Container/bucket name
	BlobName    string `json:"blob_name"`              // Full blob/object name (including path)

	// S3 fields
	Bucket string `json:"bucket,omitempty"` // S3 bucket name (alternative to container)
	Key    string `json:"key,omitempty"`    // S3 object key (alternative to blob_name)

	// Common fields
	Path string `json:"path"`          // Path within container/bucket
	URL  string `json:"url,omitempty"` // Pre-signed URL (generated dynamically, not stored)

	// Metadata for SAS generation
	ContentType string `json:"content_type,omitempty"` // Content type for proper SAS headers
}

// IsAzure returns true if this is an Azure Blob Storage reference
func (c *CloudStorageRef) IsAzure() bool {
	return c.Provider == "azure"
}

// IsS3 returns true if this is an S3 reference
func (c *CloudStorageRef) IsS3() bool {
	return c.Provider == "s3"
}

// GetStorageKey returns the appropriate storage key based on provider
func (c *CloudStorageRef) GetStorageKey() string {
	if c.IsAzure() {
		return c.BlobName
	}
	if c.IsS3() {
		return c.Key
	}
	return c.Path // fallback
}

// GetContainerName returns the appropriate container/bucket name based on provider
func (c *CloudStorageRef) GetContainerName() string {
	if c.IsAzure() {
		return c.Container
	}
	if c.IsS3() {
		return c.Bucket
	}
	return c.Container // fallback
}

// Value implements the driver.Valuer interface for CloudStorageRef
func (c *CloudStorageRef) Value() (driver.Value, error) {
	if c == nil {
		return nil, nil
	}
	return json.Marshal(c)
}

// Scan implements the sql.Scanner interface for CloudStorageRef
func (c *CloudStorageRef) Scan(value interface{}) error {
	if value == nil {
		return nil
	}
	b, ok := value.([]byte)
	if !ok {
		return errors.New("type assertion to []byte failed")
	}
	return json.Unmarshal(b, c)
}

// NewArtifact creates a new artifact with a generated UUID
func NewArtifact(jobID int64, name string, artifactType ArtifactType, contentType string, size int64) *Artifact {
	return &Artifact{
		ID:          uuid.New().String(),
		JobID:       jobID,
		Name:        name,
		Type:        artifactType,
		ContentType: contentType,
		Size:        size,
		CreatedAt:   time.Now(),
	}
}

// LogLevel represents the severity level of a log message
type LogLevel string

const (
	LogLevelDebug LogLevel = "debug"
	LogLevelInfo  LogLevel = "info"
	LogLevelWarn  LogLevel = "warn"
	LogLevelError LogLevel = "error"
)

// Storage provider constants
const (
	StorageProviderAzure = "azure"
	StorageProviderS3    = "s3"
)

// JobLog represents a log entry from a data job execution
type JobLog struct {
	ID        string                 `json:"id"`
	JobID     int64                  `json:"job_id"`
	Level     LogLevel               `json:"level"`
	Message   string                 `json:"message"`
	Timestamp time.Time              `json:"timestamp"`
	Metadata  map[string]interface{} `json:"metadata,omitempty"`
}

// NewJobLog creates a new job log entry
func NewJobLog(jobID int64, level LogLevel, message string, metadata map[string]interface{}) *JobLog {
	return &JobLog{
		ID:        uuid.New().String(),
		JobID:     jobID,
		Level:     level,
		Message:   message,
		Timestamp: time.Now(),
		Metadata:  metadata,
	}
}
