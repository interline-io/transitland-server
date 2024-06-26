package model

import (
	"context"
	"net/http"

	"github.com/interline-io/transitland-lib/tl"
	"github.com/interline-io/transitland-mw/jobs"
	"github.com/interline-io/transitland-server/internal/clock"
)

type Config struct {
	Finder             Finder
	RTFinder           RTFinder
	GbfsFinder         GbfsFinder
	Checker            Checker
	JobQueue           jobs.JobQueue
	Clock              clock.Clock
	Secrets            []tl.Secret
	ValidateLargeFiles bool
	DisableImage       bool
	RestPrefix         string
	Storage            string
	RTStorage          string
}

var finderCtxKey = &contextKey{"finderConfig"}

type contextKey struct {
	name string
}

func ForContext(ctx context.Context) Config {
	raw, ok := ctx.Value(finderCtxKey).(Config)
	if !ok {
		return Config{}
	}
	return raw
}

func WithConfig(ctx context.Context, cfg Config) context.Context {
	r := context.WithValue(ctx, finderCtxKey, cfg)
	return r
}

func AddConfig(cfg Config) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			r = r.WithContext(WithConfig(r.Context(), cfg))
			next.ServeHTTP(w, r)
		})
	}
}

func AddConfigAndPerms(cfg Config, next http.Handler) http.Handler {
	return AddPerms(cfg.Checker)(AddConfig(cfg)(next))
}
