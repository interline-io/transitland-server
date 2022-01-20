package config

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
}

// Redis and RT cache/job holder

type RTConfig struct {
	RedisURL string
}
