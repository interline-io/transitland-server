package storage

import (
	"context"
	"io"
	"time"
)

// Config holds common configuration for storage adapters.
type Config struct {
	Container string
	Prefix    string
	Expiry    time.Duration // For SAS URLs
}

// StorageAdapter defines the interface for external artifact storage.
type Adapter interface {
	Upload(ctx context.Context, key string, contentType string, reader io.Reader) error
	GetURL(ctx context.Context, key string) (string, error)
}
