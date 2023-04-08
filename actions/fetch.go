package actions

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/interline-io/transitland-lib/dmfr/fetch"
	"github.com/interline-io/transitland-lib/tl/tt"
	"github.com/interline-io/transitland-lib/tldb"
	"github.com/interline-io/transitland-server/auth"
	"github.com/interline-io/transitland-server/config"
	"github.com/interline-io/transitland-server/model"
	"google.golang.org/protobuf/proto"
)

func StaticFetch(ctx context.Context, cfg config.Config, dbf model.Finder, feedId string, feedUrl string, user auth.User) (*model.FeedVersionFetchResult, error) {
	userName := ""
	if user != nil {
		userName = user.Name()
	}
	feeds, err := dbf.FindFeeds(ctx, nil, nil, nil, &model.FeedFilter{OnestopID: &feedId})
	if err != nil {
		return nil, err
	}
	if len(feeds) == 0 {
		return nil, errors.New("feed not found")
	}
	feed := feeds[0].Feed
	atx := tldb.NewPostgresAdapterFromDBX(dbf.DBX())
	fetchOpts := fetch.Options{
		FeedID:    feed.ID,
		URLType:   "static_current",
		FeedURL:   feedUrl,
		Storage:   cfg.Storage,
		Secrets:   cfg.Secrets,
		FetchedAt: time.Now(),
		CreatedBy: tt.NewString(userName),
	}
	fv, fr, err := fetch.StaticFetch(atx, fetchOpts)
	if err != nil {
		return nil, err
	}
	mr := model.FeedVersionFetchResult{
		FoundSHA1: fr.Found,
	}
	if fr.FetchError == nil {
		mr.FeedVersion = &model.FeedVersion{FeedVersion: fv}
		mr.FetchError = nil
	} else {
		return nil, fr.FetchError
	}
}

func RTFetch(ctx context.Context, cfg config.Config, dbf model.Finder, rtf model.RTFinder, target string, feedId string, feedUrl string, urlType string) error {
	// Find feed
	rtfeeds, err := dbf.FindFeeds(ctx, nil, nil, nil, &model.FeedFilter{OnestopID: &feedId})
	if err != nil {
		return err
	}
	if len(rtfeeds) == 0 {
		return errors.New("feed not found")
	}
	// Make request
	rtfeed := rtfeeds[0].Feed
	atx := tldb.NewPostgresAdapterFromDBX(dbf.DBX())
	fetchOpts := fetch.Options{
		FeedID:    rtfeed.ID,
		URLType:   urlType,
		FeedURL:   feedUrl,
		Storage:   cfg.Storage,
		Secrets:   cfg.Secrets,
		FetchedAt: time.Now(),
	}
	rtmsg, fr, err := fetch.RTFetch(atx, fetchOpts)
	if err != nil {
		return err
	}
	if fr.FetchError != nil {
		return err
	}
	if rtmsg == nil {
		return err
	}
	// Convert back to bytes...
	rtdata, err := proto.Marshal(rtmsg)
	if err != nil {
		return errors.New("invalid rt data")
	}
	// Save to cache
	key := fmt.Sprintf("rtdata:%s:%s", target, urlType)
	return rtf.AddData(key, rtdata)
}
