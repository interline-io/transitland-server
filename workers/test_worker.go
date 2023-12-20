package workers

import (
	"context"
	"errors"

	"github.com/interline-io/transitland-mw/jobs"
)

type testOkWorker struct{}

func (w *testOkWorker) Run(ctx context.Context, job jobs.Job) error {
	job.Logger.Info().Msg("testOkWorker")
	return nil
}

type testFailWorker struct{}

func (w *testFailWorker) Run(ctx context.Context, job jobs.Job) error {
	job.Logger.Error().Msg("testFailWorker")
	return errors.New("testFailWorker")
}
