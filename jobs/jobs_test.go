package jobs

import (
	"context"
	"sync/atomic"
	"time"
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
