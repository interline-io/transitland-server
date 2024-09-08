package workers

import (
	"context"
	"errors"

	"github.com/interline-io/log"
	"github.com/interline-io/transitland-mw/jobs"
	"github.com/interline-io/transitland-server/actions"
)

type StaticFetchWorker struct {
	FeedUrl    string `json:"feed_url"`
	FeedID     string `json:"feed_id"`
	FetchEpoch int64  `json:"fetch_epoch"`
}

func (w *StaticFetchWorker) Run(ctx context.Context, _ jobs.Job) error {
	log := log.For(ctx)
	log.Info().Str("feed_id", w.FeedID).Str("feed_url", w.FeedUrl).Msg("static-fetch: started")
	if result, err := actions.StaticFetch(ctx, w.FeedID, nil, w.FeedUrl); err != nil {
		log.Error().Err(err).Msg("static-fetch: request failed")
		return err
	} else if result.FetchError != nil {
		err = errors.New(*result.FetchError)
		log.Error().Err(err).Msg("static-fetch: request failed")
		return err
	}
	log.Info().Msg("static-fetch: success")
	return nil
}
