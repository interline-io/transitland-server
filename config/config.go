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

// Job queue
type JobQueue interface {
	AddJob(Job) error
	AddWorker(func(Job) error, int) error
	Run() error
	Stop() error
}

type Job struct {
	JobType string   `json:"job_type"`
	Feed    string   `json:"feed"`
	URL     string   `json:"url"`
	Args    []string `json:"args"`
}
