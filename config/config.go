package config

import (
	"github.com/go-redis/redis/v8"
	"github.com/jmoiron/sqlx"
)

// Config is in a separate package to avoid import cycles.

type Config struct {
	GtfsDir            string
	GtfsS3Bucket       string
	ValidateLargeFiles bool
	DisableImage       bool
	RestPrefix         string
	DB                 DBConfig
	RT                 RTConfig
}

// Connection holder

type DBConfig struct {
	DBURL string
	DB    sqlx.Ext
}

// Redis and RT cache/job holder

type RTConfig struct {
	RedisURL string
	Redis    *redis.Client
}
