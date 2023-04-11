package resolvers

import (
	"context"
	"errors"
	"io"

	"github.com/99designs/gqlgen/graphql"

	"github.com/interline-io/transitland-server/actions"
	"github.com/interline-io/transitland-server/auth"
	"github.com/interline-io/transitland-server/model"
)

// mutation root

type mutationResolver struct{ *Resolver }

func (r *mutationResolver) ValidateGtfs(ctx context.Context, file *graphql.Upload, url *string, rturls []string) (*model.ValidationResult, error) {
	var src io.Reader
	if file != nil {
		src = file.File
	}
	return actions.ValidateUpload(ctx, r.cfg, src, url, rturls, auth.ForContext(ctx))
}

func (r *mutationResolver) FeedVersionFetch(ctx context.Context, file *graphql.Upload, url *string, feedId string) (*model.FeedVersionFetchResult, error) {
	// This is checked by a GraphQL directive, but we'll check again for now.
	user := auth.ForContext(ctx)
	if user == nil || !user.HasRole("admin") {
		return nil, errors.New("permission denied")
	}
	var feedSrc io.Reader
	if file != nil {
		feedSrc = file.File
	}
	feedUrl := ""
	if url != nil {
		feedUrl = *url
	}
	return actions.StaticFetch(ctx, r.cfg, r.finder, feedId, feedSrc, feedUrl, user)
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
