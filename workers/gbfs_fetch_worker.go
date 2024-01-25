package workers

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/interline-io/log"
	"github.com/interline-io/transitland-lib/tldb"
	"github.com/interline-io/transitland-mw/jobs"
	"github.com/interline-io/transitland-server/internal/gbfs"
	"github.com/interline-io/transitland-server/model"
)

type GbfsFetchWorker struct {
	Url        string `json:"url"`
	FeedID     string `json:"feed_id"`
	FetchEpoch int64  `json:"fetch_epoch"`
}

func (w *GbfsFetchWorker) Run(ctx context.Context, job jobs.Job) error {
	cfg := model.ForContext(ctx)
	log := log.For(ctx)
	log.Info().Str("feed_id", w.FeedID).Str("url", w.Url).Msg("gbfs-fetch: started")
	gfeeds, err := cfg.Finder.FindFeeds(ctx, nil, nil, nil, &model.FeedFilter{OnestopID: &w.FeedID})
	if err != nil {
		log.Error().Err(err).Msg("gbfs-fetch: error loading source feed")
		return err
	}
	if len(gfeeds) == 0 {
		log.Error().Err(err).Msg("gbfs-fetch: source feed not found")
		return errors.New("feed not found")
	}

	// Make request
	opts := gbfs.Options{}
	opts.FeedURL = gfeeds[0].URLs.GbfsAutoDiscovery
	opts.FeedID = gfeeds[0].ID
	opts.URLType = "gbfs_auto_discovery"
	opts.FetchedAt = time.Now().In(time.UTC)
	if w.Url != "" {
		opts.FeedURL = w.Url
	}
	feeds, result, err := gbfs.Fetch(
		tldb.NewPostgresAdapterFromDBX(cfg.Finder.DBX()),
		opts,
	)
	if err != nil {
		return err
	}
	if result.FetchError != nil {
		return result.FetchError
	}

	// Save to cache
	for _, feed := range feeds {
		if feed.SystemInformation != nil {
			key := fmt.Sprintf("%s:%s", w.FeedID, feed.SystemInformation.Language.Val)
			cfg.GbfsFinder.AddData(ctx, key, feed)
		}
	}
	log.Info().Msg("gbfs-fetch: success")
	return nil
}
