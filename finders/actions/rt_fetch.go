package actions

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/interline-io/transitland-lib/dmfr/fetch"
	"github.com/interline-io/transitland-lib/rt/pb"
	"github.com/interline-io/transitland-lib/tldb"
	"github.com/interline-io/transitland-server/model"
	"google.golang.org/protobuf/proto"
)

func RTFetch(ctx context.Context, target string, feedId string, feedUrl string, urlType string) error {
	cfg := model.ForContext(ctx)

	feed, err := fetchCheckFeed(ctx, feedId)
	if err != nil {
		return err
	}
	if feed == nil {
		return nil
	}

	// Prepare
	fetchOpts := fetch.Options{
		FeedID:    feed.ID,
		URLType:   urlType,
		FeedURL:   feedUrl,
		Storage:   cfg.RTStorage,
		Secrets:   cfg.Secrets,
		FetchedAt: time.Now().In(time.UTC),
	}

	// Make request
	var rtMsg *pb.FeedMessage
	var fetchErr error
	if err := tldb.NewPostgresAdapterFromDBX(cfg.Finder.DBX()).Tx(func(atx tldb.Adapter) error {
		fr, err := fetch.RTFetch(atx, fetchOpts)
		if err != nil {
			return err
		}
		rtMsg = fr.Message
		fetchErr = fr.FetchError
		return nil
	}); err != nil {
		return err
	}

	// Check result and cache
	if fetchErr != nil {
		return fetchErr
	}
	rtdata, err := proto.Marshal(rtMsg)
	if err != nil {
		return errors.New("invalid rt data")
	}
	key := fmt.Sprintf("rtdata:%s:%s", target, urlType)
	return cfg.RTFinder.AddData(key, rtdata)
}
