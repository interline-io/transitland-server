package workers

import (
	"context"

	"github.com/interline-io/log"
	"github.com/interline-io/transitland-mw/jobs"
	"github.com/interline-io/transitland-server/actions"
)

type RTFetchWorker struct {
	Target       string `json:"target"`
	Url          string `json:"url"`
	SourceType   string `json:"source_type"`
	SourceFeedID string `json:"source_feed_id"`
	FetchEpoch   int64  `json:"fetch_epoch"`
}

func (w *RTFetchWorker) Kind() string {
	return "rt-fetch"
}

func (w *RTFetchWorker) Run(ctx context.Context, job jobs.Job) error {
	log := log.For(ctx)
	log.Info().Str("target", w.Target).Str("source_feed_id", w.SourceFeedID).Str("source_type", w.SourceType).Str("url", w.Url).Msg("rt-fetch: started")
	err := actions.RTFetch(ctx, w.Target, w.SourceFeedID, w.Url, w.SourceType)
	if err != nil {
		log.Error().Err(err).Msg("rt-fetch: request failed")
		return err
	}
	log.Info().Msg("rt-fetch: success")
	return err
}
