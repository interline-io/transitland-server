package actions

import (
	"context"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"time"

	"github.com/interline-io/transitland-lib/dmfr/fetch"
	"github.com/interline-io/transitland-lib/rt/pb"
	"github.com/interline-io/transitland-lib/tl/tt"
	"github.com/interline-io/transitland-lib/tldb"
	"github.com/interline-io/transitland-server/auth/authn"
	"github.com/interline-io/transitland-server/auth/authz"
	"github.com/interline-io/transitland-server/config"
	"github.com/interline-io/transitland-server/model"
	"google.golang.org/protobuf/proto"
)

func StaticFetch(ctx context.Context, cfg config.Config, dbf model.Finder, feedId string, feedSrc io.Reader, feedUrl string, user authn.User, checker *authz.Checker) (*model.FeedVersionFetchResult, error) {
	// Check feed exists
	feeds, err := dbf.FindFeeds(ctx, nil, nil, nil, nil, &model.FeedFilter{OnestopID: &feedId})
	if err != nil {
		return nil, err
	}
	if len(feeds) == 0 {
		return nil, errors.New("feed not found")
	}
	feed := feeds[0].Feed

	// Check feed permissions
	if checker != nil {
		if check, err := checker.FeedPermissions(ctx, &authz.FeedRequest{Id: int64(feed.ID)}); err != nil {
			return nil, err
		} else if !check.Actions.CanEdit {
			return nil, authz.ErrUnauthorized
		}
	}

	// Prepare
	fetchOpts := fetch.Options{
		FeedID:    feed.ID,
		URLType:   "static_current",
		FeedURL:   feedUrl,
		Storage:   cfg.Storage,
		Secrets:   cfg.Secrets,
		FetchedAt: time.Now(),
	}
	if user != nil {
		fetchOpts.CreatedBy = tt.NewString(user.Name())
	}

	// Allow a Reader
	if feedSrc != nil {
		tmpfile, err := ioutil.TempFile("", "validator-upload")
		if err != nil {
			return nil, err
		}
		io.Copy(tmpfile, feedSrc)
		tmpfile.Close()
		defer os.Remove(tmpfile.Name())
		fetchOpts.FeedURL = tmpfile.Name()
		fetchOpts.AllowLocalFetch = true
	}

	// Make request
	mr := model.FeedVersionFetchResult{}
	db := tldb.NewPostgresAdapterFromDBX(dbf.DBX())
	if err := db.Tx(func(atx tldb.Adapter) error {
		fv, fr, err := fetch.StaticFetch(atx, fetchOpts)
		if err != nil {
			return err
		}
		mr.FoundSHA1 = fr.Found
		if fr.FetchError == nil {
			mr.FeedVersion = &model.FeedVersion{FeedVersion: fv}
			mr.FetchError = nil
		} else {
			a := fr.FetchError.Error()
			mr.FetchError = &a
		}
		return nil
	}); err != nil {
		return nil, err
	}
	return &mr, nil
}

func RTFetch(ctx context.Context, cfg config.Config, dbf model.Finder, rtf model.RTFinder, target string, feedId string, feedUrl string, urlType string, checker *authz.Checker) error {
	// Check feed exists
	rtfeeds, err := dbf.FindFeeds(ctx, nil, nil, nil, nil, &model.FeedFilter{OnestopID: &feedId})
	if err != nil {
		return err
	}
	if len(rtfeeds) == 0 {
		return errors.New("feed not found")
	}
	rtfeed := rtfeeds[0]

	// Check feed permissions
	if checker != nil {
		if check, err := checker.FeedPermissions(ctx, &authz.FeedRequest{Id: int64(rtfeed.ID)}); err != nil {
			return err
		} else if !check.Actions.CanEdit {
			return authz.ErrUnauthorized
		}
	}

	// Prepare
	fetchOpts := fetch.Options{
		FeedID:    rtfeed.ID,
		URLType:   urlType,
		FeedURL:   feedUrl,
		Storage:   cfg.RTStorage,
		Secrets:   cfg.Secrets,
		FetchedAt: time.Now(),
	}

	// Make request
	var rtMsg *pb.FeedMessage
	var fetchErr error
	db := tldb.NewPostgresAdapterFromDBX(dbf.DBX())
	if err := db.Tx(func(atx tldb.Adapter) error {
		m, fr, err := fetch.RTFetch(atx, fetchOpts)
		if err != nil {
			return err
		}
		rtMsg = m
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
	return rtf.AddData(key, rtdata)
}
