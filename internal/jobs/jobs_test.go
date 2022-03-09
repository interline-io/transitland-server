package jobs

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

var (
	feeds = []string{"BA", "SF", "AC", "CT"}
)

type testWorker struct {
	count  int
	FeedID string `json:"feed_id"`
}

func (t *testWorker) Run(ctx context.Context, job Job) error {
	time.Sleep(10 * time.Millisecond)
	t.count += 1
	return nil
}

func testGetWorker(job Job) (JobWorker, error) {
	w := testWorker{}
	// Load json
	jw, err := json.Marshal(job.JobArgs)
	if err != nil {
		return nil, err
	}
	if err := json.Unmarshal(jw, &w); err != nil {
		return nil, err
	}
	return &w, nil

}

func testJobs(t *testing.T, rtJobs JobQueue) {
	w := testWorker{}
	rtJobs.AddWorker(testGetWorker, JobOptions{}, 1)
	for _, feed := range feeds {
		rtJobs.AddJob(Job{JobType: "test", JobArgs: JobArgs{"feed_id": feed}})
	}
	go func() {
		time.Sleep(100 * time.Millisecond)
		rtJobs.Stop()
	}()
	rtJobs.Run()
	assert.Equal(t, len(feeds), w.count)
}
