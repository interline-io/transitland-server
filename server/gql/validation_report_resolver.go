package gql

import (
	"context"

	"github.com/interline-io/transitland-server/model"
)

type validationReportResolver struct{ *Resolver }

func (r *validationReportResolver) Errors(ctx context.Context, obj *model.ValidationReport, limit *int) ([]*model.ValidationReportErrorGroup, error) {
	if len(obj.Errors) > 0 {
		return sliceToPointerSlice(obj.Errors), nil
	}
	return For(ctx).ValidationReportErrorGroupsByValidationReportID.Load(ctx, model.ValidationReportErrorGroupParam{ValidationReportID: obj.ID, Limit: limit})()
}

func (r *validationReportResolver) Warnings(ctx context.Context, obj *model.ValidationReport, limit *int) ([]*model.ValidationReportErrorGroup, error) {
	if len(obj.Warnings) > 0 {
		return sliceToPointerSlice(obj.Warnings), nil
	}
	return nil, nil
}

func (r *validationReportResolver) Details(ctx context.Context, obj *model.ValidationReport) (*model.ValidationReportDetails, error) {
	return &obj.Details, nil
}

type validationReportErrorGroupResolver struct{ *Resolver }

func (r *validationReportErrorGroupResolver) Errors(ctx context.Context, obj *model.ValidationReportErrorGroup, limit *int) ([]*model.ValidationReportError, error) {
	if len(obj.Errors) > 0 {
		return obj.Errors, nil
	}
	ret, err := For(ctx).ValidationReportErrorExemplarsByValidationReportErrorGroupID.Load(ctx, model.ValidationReportErrorExemplarParam{ValidationReportGroupID: obj.ID, Limit: limit})()
	if err != nil {
		return nil, err
	}
	for _, r := range ret {
		r.ErrorCode = obj.ErrorCode
		r.ErrorType = obj.ErrorType
		r.Field = obj.Field
		r.Filename = obj.Filename
	}
	return ret, nil
}

func sliceToPointerSlice[T any](a []T) []*T {
	var ret []*T
	for _, a := range a {
		a := a
		ret = append(ret, &a)
	}
	return ret
}
