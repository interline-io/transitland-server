package actions

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/interline-io/log"
	"github.com/interline-io/transitland-lib/tldb"
	"github.com/interline-io/transitland-server/internal/gbfs"
	"github.com/interline-io/transitland-server/model"
)

func GBFSFetch(ctx context.Context, feedId string, feedUrl string) error {
	cfg := model.ForContext(ctx)
	log := log.For(ctx)
	gfeeds, err := cfg.Finder.FindFeeds(ctx, nil, nil, nil, &model.FeedFilter{OnestopID: &feedId})
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
	if feedUrl != "" {
		opts.FeedURL = feedUrl
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
			key := fmt.Sprintf("%s:%s", feedId, feed.SystemInformation.Language.Val)
			cfg.GbfsFinder.AddData(ctx, key, feed)
		}
	}
	return nil
}
