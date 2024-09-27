package actions

import (
	"context"
	"io"
	"io/ioutil"
	"os"
	"time"

	"github.com/interline-io/transitland-lib/dmfr/fetch"
	"github.com/interline-io/transitland-lib/tl/tt"
	"github.com/interline-io/transitland-lib/tldb"
	"github.com/interline-io/transitland-mw/auth/authn"
	"github.com/interline-io/transitland-server/model"
)

func StaticFetch(ctx context.Context, feedId string, feedSrc io.Reader, feedUrl string) (*model.FeedVersionFetchResult, error) {
	cfg := model.ForContext(ctx)
	dbf := cfg.Finder

	urlType := "static_current"
	feed, err := fetchCheckFeed(ctx, feedId)
	if err != nil {
		return nil, err
	}
	if feed == nil {
		return nil, nil
	}

	// Prepare
	fetchOpts := fetch.Options{
		FeedID:        feed.ID,
		URLType:       urlType,
		FeedURL:       feedUrl,
		Storage:       cfg.Storage,
		Secrets:       cfg.Secrets,
		FetchedAt:     time.Now().In(time.UTC),
		AllowFTPFetch: true,
	}

	if user := authn.ForContext(ctx); user != nil {
		fetchOpts.CreatedBy = tt.NewString(user.ID())
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
		fr, err := fetch.StaticFetch(atx, fetchOpts)
		if err != nil {
			return err
		}
		mr.FoundSha1 = fr.Found
		if fr.FetchError != nil {
			a := fr.FetchError.Error()
			mr.FetchError = &a
		} else if fr.FeedVersion != nil {
			mr.FeedVersion = &model.FeedVersion{FeedVersion: *fr.FeedVersion}
			mr.FetchError = nil
		}
		return nil
	}); err != nil {
		return nil, err
	}
	return &mr, nil
}
