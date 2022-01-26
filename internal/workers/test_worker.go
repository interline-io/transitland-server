package workers

import (
	"context"
	"fmt"

	"github.com/interline-io/transitland-server/internal/jobs"
)

type testWorker struct{}

func (w *testWorker) Run(ctx context.Context, job jobs.Job) error {
	fmt.Println("test worker")
	return nil
}
