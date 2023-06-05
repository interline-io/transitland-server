package workers

import (
	"context"
	"errors"

	"github.com/interline-io/transitland-server/actions"
	"github.com/interline-io/transitland-server/internal/jobs"
)

type StaticFetchWorker struct {
	FeedUrl string `json:"feed_url"`
	FeedID  string `json:"feed_id"`
}

func (w *StaticFetchWorker) Run(ctx context.Context, job jobs.Job) error {
	log := job.Opts.Logger.With().Str("feed_id", w.FeedID).Str("feed_url", w.FeedUrl).Logger()
	if result, err := actions.StaticFetch(ctx, job.Opts.Config, job.Opts.Finder, w.FeedID, nil, w.FeedUrl, nil, nil); err != nil {
		log.Error().Err(err).Msg("staticfetch worker: request failed")
		return err
	} else if result.FetchError != nil {
		err = errors.New(*result.FetchError)
		log.Error().Err(err).Msg("staticfetch worker: request failed")
		return err
	}
	log.Info().Msg("staticfetch worker: success")
	return nil
}
