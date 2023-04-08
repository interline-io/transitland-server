package resolvers

import (
	"context"
	"errors"
	"io"
	"io/ioutil"
	"os"
	"time"

	"github.com/99designs/gqlgen/graphql"
	"github.com/interline-io/transitland-lib/log"

	"github.com/interline-io/transitland-lib/dmfr/fetch"
	"github.com/interline-io/transitland-lib/tl"
	"github.com/interline-io/transitland-lib/tl/tt"
	"github.com/interline-io/transitland-lib/tldb"
	"github.com/interline-io/transitland-server/actions"
	"github.com/interline-io/transitland-server/auth"
	"github.com/interline-io/transitland-server/config"
	"github.com/interline-io/transitland-server/model"
)

// mutation root

type mutationResolver struct{ *Resolver }

func (r *mutationResolver) ValidateGtfs(ctx context.Context, file *graphql.Upload, url *string, rturls []string) (*model.ValidationResult, error) {
	var src io.Reader
	if file != nil {
		src = file.File
	}
	return actions.ValidateUpload(r.cfg, src, url, rturls, auth.ForContext(ctx))
}

func (r *mutationResolver) FeedVersionFetch(ctx context.Context, file *graphql.Upload, url *string, feed string) (*model.FeedVersionFetchResult, error) {
	var src io.Reader
	if file != nil {
		src = file.File
	}
	return Fetch(r.cfg, r.finder, src, url, feed, auth.ForContext(ctx))
}

func (r *mutationResolver) FeedVersionImport(ctx context.Context, sha1 string) (*model.FeedVersionImportResult, error) {
	return nil, errors.New("temporarily unavailable")
}

func (r *mutationResolver) FeedVersionUpdate(ctx context.Context, id int, values model.FeedVersionSetInput) (*model.FeedVersion, error) {
	return nil, errors.New("temporarily unavailable")
}

func (r *mutationResolver) FeedVersionUnimport(ctx context.Context, id int) (*model.FeedVersionUnimportResult, error) {
	return nil, errors.New("temporarily unavailable")
}

func (r *mutationResolver) FeedVersionDelete(ctx context.Context, id int) (*model.FeedVersionDeleteResult, error) {
	return nil, errors.New("temporarily unavailable")
}

// Fetch adds a feed version to the database.
func Fetch(cfg config.Config, finder model.Finder, src io.Reader, feedURL *string, feedId string, user auth.User) (*model.FeedVersionFetchResult, error) {
	if user == nil {
		return nil, errors.New("no user")
	}
	// Find feed
	// feeds, err := cfg.Finder.FindFeeds(nil, nil, nil, &model.FeedFilter{OnestopID: &feed})
	var feeds []tl.Feed
	atx := tldb.NewPostgresAdapterFromDBX(finder.DBX())
	err := atx.Select(&feeds, "select * from current_feeds where onestop_id = ?", feedId)
	if err != nil {
		log.Error().Err(err).Msg("fetch mutation: error loading source feed")
		return nil, err
	}
	if len(feeds) == 0 {
		log.Error().Err(err).Msg("fetch mutation: source feed not found")
		return nil, errors.New("feed not found")
	}
	feed := feeds[0]
	// Prepare request
	opts := fetch.Options{
		URLType:   "manual",
		FetchedAt: time.Now(),
		FeedID:    feed.ID,
		Storage:   cfg.Storage,
		CreatedBy: tt.NewString(user.Name()),
	}
	if src != nil {
		// Prepare reader
		tmpfile, err := ioutil.TempFile("", "validator-upload")
		if err != nil {
			// This should result in a failed request
			return nil, err
		}
		io.Copy(tmpfile, src)
		tmpfile.Close()
		defer os.Remove(tmpfile.Name())
		opts.FeedURL = tmpfile.Name()
	} else if feedURL != nil {
		opts.FeedURL = *feedURL
	}
	// Make request
	fv, fr, err := fetch.StaticFetch(atx, opts)
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
	return &mr, nil
}
