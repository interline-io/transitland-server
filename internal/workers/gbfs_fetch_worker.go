package workers

import (
	"context"
	"errors"
	"fmt"

	"github.com/interline-io/transitland-server/internal/gbfs"
	"github.com/interline-io/transitland-server/internal/jobs"
	"github.com/interline-io/transitland-server/model"
)

type GbfsFetchWorker struct {
	Target       string `json:"target"`
	Url          string `json:"url"`
	SourceType   string `json:"source_type"`
	SourceFeedID string `json:"source_feed_id"`
}

func (w *GbfsFetchWorker) Run(ctx context.Context, job jobs.Job) error {
	log := job.Opts.Logger.With().Str("target", w.Target).Str("source_feed_id", w.SourceFeedID).Str("source_type", w.SourceType).Str("url", w.Url).Logger()
	gfeeds, err := job.Opts.Finder.FindFeeds(ctx, nil, nil, nil, nil, &model.FeedFilter{OnestopID: &w.SourceFeedID})
	if err != nil {
		log.Error().Err(err).Msg("gbfsfetch worker: error loading source feed")
		return err
	}
	if len(gfeeds) == 0 {
		log.Error().Err(err).Msg("gbfsfetch worker: source feed not found")
		return errors.New("feed not found")
	}
	// Make request
	opts := gbfs.Options{}
	opts.FeedURL = gfeeds[0].URLs.GbfsAutoDiscovery
	if w.Url != "" {
		opts.FeedURL = w.Url
	}
	feeds, result, err := gbfs.Fetch(opts)
	if err != nil {
		return err
	}
	if result.FetchError != nil {
		return result.FetchError
	}
	// Save to cache
	for _, feed := range feeds {
		key := fmt.Sprintf("%s:%s", w.SourceFeedID, feed.SystemInformation.Language.Val)
		job.Opts.GbfsFinder.AddData(ctx, key, feed)
	}
	log.Info().Msg("gbfs fetch worker: success")
	return nil
}
