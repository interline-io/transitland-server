package dbfinder

import (
	"context"

	sq "github.com/Masterminds/squirrel"
	"github.com/interline-io/transitland-dbutil/dbutil"
	"github.com/interline-io/transitland-server/model"
)

func (f *Finder) ValidationReportsByFeedVersionID(ctx context.Context, params []model.ValidationReportParam) ([][]*model.ValidationReport, []error) {
	return paramGroupQuery(
		params,
		func(p model.ValidationReportParam) (int, *model.ValidationReportFilter, *int) {
			return p.FeedVersionID, p.Where, p.Limit
		},
		func(keys []int, where *model.ValidationReportFilter, limit *int) ([]*model.ValidationReport, error) {
			q := sq.StatementBuilder.
				Select("*").
				From("tl_validation_reports").
				Limit(checkLimit(limit)).
				OrderBy("tl_validation_reports.created_at desc, tl_validation_reports.id desc")
			if where != nil {
				if len(where.ReportIds) > 0 {
					q = q.Where(In("tl_validation_reports.id", where.ReportIds))
				}
				if where.Success != nil {
					q = q.Where(sq.Eq{"success": where.Success})
				}
				if where.Validator != nil {
					q = q.Where(sq.Eq{"validator": where.Validator})
				}
				if where.ValidatorVersion != nil {
					q = q.Where(sq.Eq{"validator_version": where.ValidatorVersion})
				}
				if where.IncludesRt != nil {
					q = q.Where(sq.Eq{"includes_rt": where.IncludesRt})
				}
				if where.IncludesStatic != nil {
					q = q.Where(sq.Eq{"includes_static": where.IncludesStatic})
				}
			}
			var ents []*model.ValidationReport
			err := dbutil.Select(ctx,
				f.db,
				lateralWrap(
					q,
					"feed_versions",
					"id",
					"tl_validation_reports",
					"feed_version_id",
					keys,
				),
				&ents,
			)
			return ents, err
		},
		func(ent *model.ValidationReport) int { return ent.FeedVersionID },
	)
}

func (f *Finder) ValidationReportErrorGroupsByValidationReportID(ctx context.Context, params []model.ValidationReportErrorGroupParam) ([][]*model.ValidationReportErrorGroup, []error) {
	return paramGroupQuery(
		params,
		func(p model.ValidationReportErrorGroupParam) (int, bool, *int) {
			return p.ValidationReportID, false, p.Limit
		},
		func(keys []int, where bool, limit *int) ([]*model.ValidationReportErrorGroup, error) {
			var ents []*model.ValidationReportErrorGroup
			err := dbutil.Select(ctx,
				f.db,
				lateralWrap(
					quickSelect("tl_validation_report_error_groups", limit, nil, nil),
					"tl_validation_reports",
					"id",
					"tl_validation_report_error_groups",
					"validation_report_id",
					keys,
				),
				&ents,
			)
			return ents, err
		},
		func(ent *model.ValidationReportErrorGroup) int { return ent.ValidationReportID },
	)
}

func (f *Finder) ValidationReportErrorExemplarsByValidationReportErrorGroupID(ctx context.Context, params []model.ValidationReportErrorExemplarParam) ([][]*model.ValidationReportError, []error) {
	return paramGroupQuery(
		params,
		func(p model.ValidationReportErrorExemplarParam) (int, bool, *int) {
			return p.ValidationReportGroupID, false, p.Limit
		},
		func(keys []int, where bool, limit *int) ([]*model.ValidationReportError, error) {
			var ents []*model.ValidationReportError
			err := dbutil.Select(ctx,
				f.db,
				lateralWrap(
					quickSelect("tl_validation_report_error_exemplars", limit, nil, nil),
					"tl_validation_report_error_groups",
					"id",
					"tl_validation_report_error_exemplars",
					"validation_report_error_group_id",
					keys,
				),
				&ents,
			)
			return ents, err
		},
		func(ent *model.ValidationReportError) int { return ent.ValidationReportErrorGroupID },
	)
}
