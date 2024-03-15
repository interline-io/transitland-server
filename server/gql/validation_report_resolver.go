package gql

import (
	"context"

	"github.com/interline-io/transitland-server/model"
)

type validationReportResolver struct{ *Resolver }

func (r *validationReportResolver) Errors(ctx context.Context, obj *model.ValidationReport, limit *int) ([]*model.ValidationReportErrorGroup, error) {
	if len(obj.Errors) > 0 {
		return obj.Errors, nil
	}
	return For(ctx).ValidationReportErrorGroupsByValidationReportID.Load(ctx, model.ValidationReportErrorGroupParam{ValidationReportID: obj.ID, Limit: limit})()
}

func (r *validationReportResolver) Warnings(ctx context.Context, obj *model.ValidationReport, limit *int) ([]*model.ValidationReportErrorGroup, error) {
	if len(obj.Warnings) > 0 {
		return obj.Warnings, nil
	}
	return nil, nil
}

func (r *validationReportResolver) Details(ctx context.Context, obj *model.ValidationReport) (*model.ValidationReportDetails, error) {
	return obj.Details, nil
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
		r.GroupKey = obj.GroupKey
		r.ErrorCode = obj.ErrorCode
		r.ErrorType = obj.ErrorType
		r.Field = obj.Field
		r.Filename = obj.Filename
	}
	return ret, nil
}

type validationReportErrorResolver struct{ *Resolver }

func (r *validationReportErrorResolver) EntityJSON(ctx context.Context, obj *model.ValidationReportError) (map[string]interface{}, error) {
	return obj.EntityJSON.Val, nil
}
