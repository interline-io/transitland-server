package workers

import (
	"context"
	"errors"

	"github.com/interline-io/log"
	"github.com/interline-io/transitland-jobs/jobs"
)

type testOkWorker struct{}

func (w *testOkWorker) Run(ctx context.Context, job jobs.Job) error {
	log.For(ctx).Info().Msg("testOkWorker")
	return nil
}

func (w *testOkWorker) Kind() string {
	return "test-ok"
}

type testFailWorker struct{}

func (w *testFailWorker) Run(ctx context.Context, job jobs.Job) error {
	log.For(ctx).Error().Msg("testFailWorker")
	return errors.New("testFailWorker")
}

func (w *testFailWorker) Kind() string {
	return "test-fail"
}
