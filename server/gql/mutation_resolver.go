package gql

import (
	"context"
	"errors"
	"io"

	"github.com/99designs/gqlgen/graphql"

	"github.com/interline-io/transitland-mw/auth/authn"
	"github.com/interline-io/transitland-server/actions"
	"github.com/interline-io/transitland-server/model"
)

// mutation root

type mutationResolver struct{ *Resolver }

func (r *mutationResolver) ValidateGtfs(ctx context.Context, file *graphql.Upload, url *string, rturls []string) (*model.ValidationResult, error) {
	var src io.Reader
	if file != nil {
		src = file.File
	}
	return actions.ValidateUpload(ctx, r.cfg, src, url, rturls, authn.ForContext(ctx))
}

func (r *mutationResolver) FeedVersionFetch(ctx context.Context, file *graphql.Upload, url *string, feedId string) (*model.FeedVersionFetchResult, error) {
	var feedSrc io.Reader
	if file != nil {
		feedSrc = file.File
	}
	feedUrl := ""
	if url != nil {
		feedUrl = *url
	}
	return actions.StaticFetch(ctx, r.cfg, r.finder, feedId, feedSrc, feedUrl, authn.ForContext(ctx), r.authzChecker)
}

func (r *mutationResolver) FeedVersionImport(ctx context.Context, fvid int) (*model.FeedVersionImportResult, error) {
	return actions.FeedVersionImport(ctx, r.cfg, r.finder, r.authzChecker, authn.ForContext(ctx), fvid)
}

func (r *mutationResolver) FeedVersionUnimport(ctx context.Context, fvid int) (*model.FeedVersionUnimportResult, error) {
	return actions.FeedVersionUnimport(ctx, r.cfg, r.finder, r.authzChecker, authn.ForContext(ctx), fvid)
}

func (r *mutationResolver) FeedVersionUpdate(ctx context.Context, fvid int, values model.FeedVersionSetInput) (*model.FeedVersion, error) {
	err := actions.FeedVersionUpdate(ctx, r.cfg, r.finder, r.authzChecker, authn.ForContext(ctx), fvid, values)
	return nil, err
}

func (r *mutationResolver) FeedVersionDelete(ctx context.Context, id int) (*model.FeedVersionDeleteResult, error) {
	return nil, errors.New("temporarily unavailable")
}
