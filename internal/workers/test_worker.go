package workers

import (
	"context"

	"github.com/interline-io/transitland-server/internal/jobs"
)

type testWorker struct{}

func (w *testWorker) Run(ctx context.Context, job jobs.Job) error {
	return nil
}
