package gql

import (
	"context"
	"errors"
	"io"

	"github.com/99designs/gqlgen/graphql"

	"github.com/interline-io/transitland-server/actions"
	"github.com/interline-io/transitland-server/model"
)

// mutation root

type mutationResolver struct{ *Resolver }

func (r *mutationResolver) ValidateGtfs(ctx context.Context, file *graphql.Upload, url *string, rturls []string) (*model.ValidationReport, error) {
	var src io.Reader
	if file != nil {
		src = file.File
	}
	return actions.ValidateUpload(ctx, src, url, rturls)
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
	return actions.StaticFetch(ctx, feedId, feedSrc, feedUrl)
}

func (r *mutationResolver) FeedVersionImport(ctx context.Context, fvid int) (*model.FeedVersionImportResult, error) {
	return model.ForContext(ctx).Finder.FeedVersionImport(ctx, fvid)
}

func (r *mutationResolver) FeedVersionUnimport(ctx context.Context, fvid int) (*model.FeedVersionUnimportResult, error) {
	return model.ForContext(ctx).Finder.FeedVersionUnimport(ctx, fvid)
}

func (r *mutationResolver) FeedVersionUpdate(ctx context.Context, values model.FeedVersionSetInput) (*model.FeedVersion, error) {
	finder := model.ForContext(ctx).Finder
	entId, err := finder.FeedVersionUpdate(ctx, values)
	if err != nil {
		return nil, err
	}
	ents, errs := finder.FeedVersionsByID(ctx, []int{entId})
	return first(errs, ents)
}

func (r *mutationResolver) FeedVersionDelete(ctx context.Context, id int) (*model.FeedVersionDeleteResult, error) {
	return nil, errors.New("temporarily unavailable")
}
