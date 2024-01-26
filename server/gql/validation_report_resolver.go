package gql

import (
	"context"

	"github.com/interline-io/transitland-server/model"
)

type validationReportResolver struct{ *Resolver }

func (r *validationReportResolver) Errors(ctx context.Context, obj *model.ValidationReport) ([]*model.ValidationReportErrorGroup, error) {
	return For(ctx).ValidationReportErrorGroupsByValidationReportID.Load(ctx, obj.ID)()
}

func (r *validationReportResolver) Warnings(ctx context.Context, obj *model.ValidationReport) ([]*model.ValidationReportErrorGroup, error) {
	return nil, nil
}

func (r *validationReportResolver) Details(ctx context.Context, obj *model.ValidationReport) (*model.ValidationReportDetails, error) {
	return &obj.Details, nil
}

type validationReportErrorGroupResolver struct{ *Resolver }

func (r *validationReportErrorGroupResolver) Errors(ctx context.Context, obj *model.ValidationReportErrorGroup) ([]*model.ValidationReportError, error) {
	ret, err := For(ctx).ValidationReportErrorExemplarsByValidationReportErrorGroupID.Load(ctx, obj.ID)()
	if err != nil {
		return nil, err
	}
	for _, r := range ret {
		r := r
		r.ErrorCode = obj.ErrorCode
		r.ErrorType = obj.ErrorType
		r.Field = obj.Field
		r.Filename = obj.Filename
	}
	return ret, nil
}
