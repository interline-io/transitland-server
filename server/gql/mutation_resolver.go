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
	return actions.FeedVersionImport(ctx, fvid)
}

func (r *mutationResolver) FeedVersionUnimport(ctx context.Context, fvid int) (*model.FeedVersionUnimportResult, error) {
	return actions.FeedVersionUnimport(ctx, fvid)
}

func (r *mutationResolver) FeedVersionUpdate(ctx context.Context, fvid int, values model.FeedVersionSetInput) (*model.FeedVersion, error) {
	err := actions.FeedVersionUpdate(ctx, fvid, values)
	return nil, err
}

func (r *mutationResolver) FeedVersionDelete(ctx context.Context, id int) (*model.FeedVersionDeleteResult, error) {
	return nil, errors.New("temporarily unavailable")
}

// Entity editing
func (r *mutationResolver) CreateStop(ctx context.Context, obj model.StopInput) (*model.Stop, error) {
	finder := model.ForContext(ctx).Finder
	entId, err := finder.CreateStop(ctx, obj)
	if err != nil {
		return nil, err
	}
	ents, errs := finder.StopsByID(ctx, []int{entId})
	return first(errs, ents)
}

func (r *mutationResolver) UpdateStop(ctx context.Context, obj model.StopInput) (*model.Stop, error) {
	finder := model.ForContext(ctx).Finder
	entId, err := finder.UpdateStop(ctx, obj)
	if err != nil {
		return nil, err
	}
	ents, errs := finder.StopsByID(ctx, []int{entId})
	return first(errs, ents)
}

func (r *mutationResolver) DeleteStop(ctx context.Context, id int) (*model.DeleteResult, error) {
	finder := model.ForContext(ctx).Finder
	if err := finder.DeleteStop(ctx, id); err != nil {
		return nil, err
	}
	return &model.DeleteResult{ID: id}, nil
}

func (r *mutationResolver) CreateLevel(ctx context.Context, obj model.LevelInput) (*model.Level, error) {
	finder := model.ForContext(ctx).Finder
	entId, err := finder.CreateLevel(ctx, obj)
	if err != nil {
		return nil, err
	}
	ents, errs := finder.LevelsByID(ctx, []int{entId})
	return first(errs, ents)
}

func (r *mutationResolver) UpdateLevel(ctx context.Context, obj model.LevelInput) (*model.Level, error) {
	finder := model.ForContext(ctx).Finder
	entId, err := finder.UpdateLevel(ctx, obj)
	if err != nil {
		return nil, err
	}
	ents, errs := finder.LevelsByID(ctx, []int{entId})
	return first(errs, ents)
}

func (r *mutationResolver) DeleteLevel(ctx context.Context, id int) (*model.DeleteResult, error) {
	finder := model.ForContext(ctx).Finder
	if err := finder.DeleteLevel(ctx, id); err != nil {
		return nil, err
	}
	return &model.DeleteResult{ID: id}, nil
}

func (r *mutationResolver) CreatePathway(ctx context.Context, obj model.PathwayInput) (*model.Pathway, error) {
	finder := model.ForContext(ctx).Finder
	entId, err := finder.CreatePathway(ctx, obj)
	if err != nil {
		return nil, err
	}
	ents, errs := finder.PathwaysByID(ctx, []int{entId})
	return first(errs, ents)
}

func (r *mutationResolver) UpdatePathway(ctx context.Context, obj model.PathwayInput) (*model.Pathway, error) {
	finder := model.ForContext(ctx).Finder
	entId, err := finder.UpdatePathway(ctx, obj)
	if err != nil {
		return nil, err
	}
	ents, errs := finder.PathwaysByID(ctx, []int{entId})
	return first(errs, ents)
}

func (r *mutationResolver) DeletePathway(ctx context.Context, id int) (*model.DeleteResult, error) {
	finder := model.ForContext(ctx).Finder
	if err := finder.DeletePathway(ctx, id); err != nil {
		return nil, err
	}
	return &model.DeleteResult{ID: id}, nil
}

func first[T any](errs []error, v []T) (T, error) {
	var ret T
	if len(errs) > 0 {
		return ret, errs[0]
	}
	if len(v) == 0 {
		return ret, errors.New("not found")
	}
	return v[0], nil
}
