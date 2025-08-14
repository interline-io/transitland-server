package storage

import (
	"context"
	"io"
)

// NoOpAdapter is a storage adapter that doesn't actually store anything.
// Useful for development or when only using inline JSONB storage.
type NoOpAdapter struct{}

// NewNoOpAdapter creates a new no-op storage adapter
func NewNoOpAdapter() *NoOpAdapter {
	return &NoOpAdapter{}
}

// Upload does nothing and always succeeds
func (a *NoOpAdapter) Upload(ctx context.Context, key string, contentType string, reader io.Reader) error {
	// Read and discard the data
	_, err := io.Copy(io.Discard, reader)
	return err
}

// GetURL returns an empty string since nothing is stored
func (a *NoOpAdapter) GetURL(ctx context.Context, key string) (string, error) {
	return "", nil
}
