package workers

import (
	"context"
	"errors"

	"github.com/interline-io/log"
	"github.com/interline-io/transitland-mw/jobs"
)

type testOkWorker struct{}

func (w *testOkWorker) Run(ctx context.Context, _ jobs.Job) error {
	log.For(ctx).Info().Msg("testOkWorker")
	return nil
}

type testFailWorker struct{}

func (w *testFailWorker) Run(ctx context.Context, _ jobs.Job) error {
	log.For(ctx).Error().Msg("testFailWorker")
	return errors.New("testFailWorker")
}
