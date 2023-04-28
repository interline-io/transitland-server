package config

import (
	"github.com/interline-io/transitland-lib/tl"
	"github.com/interline-io/transitland-server/internal/clock"
)

// Config is in a separate package to avoid import cycles.

type Config struct {
	Storage            string
	RTStorage          string
	ValidateLargeFiles bool
	DisableImage       bool
	RestPrefix         string
	DBURL              string
	RedisURL           string
	Clock              clock.Clock
	Secrets            []tl.Secret
}
