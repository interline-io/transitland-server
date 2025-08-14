package jobserver

import (
	"context"
	"net/http"
	"sync"
)

// JobAuthChecker defines the interface for job access control
type JobAuthChecker interface {
	// CheckJobAccess checks if the user has access to the job
	CheckJobAccess(ctx context.Context, jobID int64) bool
}

// JobAuthMiddleware defines the interface for job auth middleware
type JobAuthMiddleware interface {
	// Middleware returns an http.Handler middleware
	Middleware(next http.Handler) http.Handler
}

// JobAuthProvider defines the interface for job auth providers
type JobAuthProvider interface {
	JobAuthChecker
	JobAuthMiddleware
}

var (
	authProviderLock sync.Mutex
	authProvider     JobAuthProvider
)

// RegisterAuthProvider registers a job auth provider
func RegisterAuthProvider(provider JobAuthProvider) {
	authProviderLock.Lock()
	defer authProviderLock.Unlock()
	authProvider = provider
}

// GetAuthProvider returns the registered job auth provider
func GetAuthProvider() JobAuthProvider {
	authProviderLock.Lock()
	defer authProviderLock.Unlock()
	return authProvider
}

// DefaultAuthProvider is a no-op auth provider that allows all access
type DefaultAuthProvider struct{}

func (p *DefaultAuthProvider) CheckJobAccess(ctx context.Context, jobID int64) bool {
	return true
}

func (p *DefaultAuthProvider) Middleware(next http.Handler) http.Handler {
	return next
}

func init() {
	// Register default provider
	RegisterAuthProvider(&DefaultAuthProvider{})
}
