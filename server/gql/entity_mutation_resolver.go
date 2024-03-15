package gql

import (
	"context"
	"errors"

	"github.com/interline-io/transitland-server/model"
)

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
