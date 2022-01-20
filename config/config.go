package config

// Config is in a separate package to avoid import cycles.

type Config struct {
	GtfsDir            string
	GtfsS3Bucket       string
	ValidateLargeFiles bool
	DisableImage       bool
	RestPrefix         string
	DBURL              string
	RedisURL           string
}
