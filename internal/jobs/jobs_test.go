package jobs

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

var (
	feeds = []string{"BA", "SF", "AC", "CT"}
)

type testWorker struct {
	count int
}

func (t *testWorker) Run(ctx context.Context, job Job) error {
	t.count += 1
	return nil
}

func testJobs(t *testing.T, rtJobs JobQueue) {
	w := testWorker{}
	gw := func(Job) (JobWorker, error) { return &w, nil }
	rtJobs.AddWorker(gw, JobOptions{}, 1)
	for _, feed := range feeds {
		rtJobs.AddJob(Job{JobType: "test", Args: []string{feed}})
	}
	go func() {
		time.Sleep(100 * time.Millisecond)
		rtJobs.Stop()
	}()
	rtJobs.Run()
	assert.Equal(t, len(feeds), w.count)
}
