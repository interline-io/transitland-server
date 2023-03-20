package jobs

import (
	"context"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

var (
	feeds = []string{"BA", "SF", "AC", "CT"}
)

type testWorker struct {
	count *int64
}

func (t *testWorker) Run(ctx context.Context, job Job) error {
	time.Sleep(10 * time.Millisecond)
	atomic.AddInt64(t.count, 1)
	return nil
}

func testJobs(t *testing.T, rtJobs JobQueue) {
	// Ugly :(
	count := int64(0)
	testGetWorker := func(job Job) (JobWorker, error) {
		w := testWorker{count: &count}
		return &w, nil
	}
	rtJobs.AddWorker(testGetWorker, JobOptions{}, 1)
	for _, feed := range feeds {
		rtJobs.AddJob(Job{JobType: "test", JobArgs: JobArgs{"feed_id": feed}})
	}
	go func() {
		time.Sleep(100 * time.Millisecond)
		rtJobs.Stop()
	}()
	rtJobs.Run()
	assert.Equal(t, len(feeds), int(count))
}
