package rtcache

import (
	"time"

	"github.com/interline-io/transitland-server/config"
)

var (
	pulsarConnectionTimeout = 30 * time.Second
	JobSchema               = `
	{
		"type": "record",
		"name": "Job",
		"namespace": "test",
		"fields": [
		{
			"name": "job_type",
			"type": "string"
		},
		{
			"name": "feed",
			"type": "string"
		}, {
			"name": "url",
			"type": "string"
		}]
	}`
)

// Refactoring
type Cache = config.Cache
type JobQueue = config.JobQueue
type Job = config.Job
