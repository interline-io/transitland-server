package workers

import (
	"context"

	"github.com/interline-io/transitland-server/actions"
	"github.com/interline-io/transitland-server/internal/jobs"
)

type RTFetchWorker struct {
	Target       string `json:"target"`
	Url          string `json:"url"`
	SourceType   string `json:"source_type"`
	SourceFeedID string `json:"source_feed_id"`
}

func (w *RTFetchWorker) Run(ctx context.Context, job jobs.Job) error {
	log := job.Opts.Logger.With().Str("target", w.Target).Str("source_feed_id", w.SourceFeedID).Str("source_type", w.SourceType).Str("url", w.Url).Logger()
	err := actions.RTFetch(ctx, job.Opts.Config, job.Opts.Finder, job.Opts.RTFinder, w.Target, w.SourceFeedID, w.Url, w.SourceType, nil)
	if err != nil {
		log.Error().Err(err).Msg("rtfetch worker: request failed")
		return err
	}
	log.Info().Msg("rtfetch worker: success")
	return err
}
