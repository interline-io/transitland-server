package model

import (
	"context"
	"net/http"

	"github.com/interline-io/transitland-lib/tl"
	"github.com/interline-io/transitland-server/internal/clock"
	"github.com/rs/zerolog"
)

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

func WithConfig(ctx context.Context, fs Config) context.Context {
	r := context.WithValue(ctx, finderCtxKey, fs)
	return r
}

func AddConfig(te Config) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			r = r.WithContext(WithConfig(r.Context(), te))
			next.ServeHTTP(w, r)
		})
	}
}

type Config struct {
	Finder             Finder
	RTFinder           RTFinder
	GbfsFinder         GbfsFinder
	Checker            Checker
	Clock              clock.Clock
	Secrets            []tl.Secret
	ValidateLargeFiles bool
	Storage            string
	RTStorage          string
	Logger             zerolog.Logger
}
