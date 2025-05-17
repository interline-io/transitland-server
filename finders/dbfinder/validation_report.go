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

func (f *Finder) ValidationReportErrorGroupsByValidationReportIDs(ctx context.Context, limit *int, keys []int) ([]*model.ValidationReportErrorGroup, error) {
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
}

func (f *Finder) ValidationReportErrorExemplarsByValidationReportErrorGroupIDs(ctx context.Context, limit *int, keys []int) (ents []*model.ValidationReportError, err error) {
	err = dbutil.Select(ctx,
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
}
