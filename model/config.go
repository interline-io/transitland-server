package model

import (
	"context"
	"fmt"
	"net/http"

	"github.com/interline-io/transitland-lib/tl"
	"github.com/interline-io/transitland-server/internal/clock"
	"github.com/rs/zerolog"
)

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
			fmt.Println("CONFIG 1")
			r = r.WithContext(WithConfig(r.Context(), cfg))
			next.ServeHTTP(w, r)
		})
	}
}
