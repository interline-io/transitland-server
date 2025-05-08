package dbfinder

import (
	"context"

	"github.com/interline-io/transitland-dbutil/dbutil"
	"github.com/interline-io/transitland-server/model"
)

func (f *Finder) CalendarsByID(ctx context.Context, ids []int) ([]*model.Calendar, []error) {
	var ents []*model.Calendar
	err := dbutil.Select(ctx,
		f.db,
		quickSelect("gtfs_calendars", nil, nil, ids),
		&ents,
	)
	if err != nil {
		return nil, logExtendErr(ctx, len(ids), err)
	}
	return arrangeBy(ids, ents, func(ent *model.Calendar) int { return ent.ID }), nil
}

func (f *Finder) CalendarDatesByServiceID(ctx context.Context, params []model.CalendarDateParam) ([][]*model.CalendarDate, []error) {
	return paramGroupQuery(
		params,
		func(p model.CalendarDateParam) (int, *model.CalendarDateFilter, *int) {
			return p.ServiceID, p.Where, p.Limit
		},
		func(keys []int, where *model.CalendarDateFilter, limit *int) (ents []*model.CalendarDate, err error) {
			err = dbutil.Select(ctx,
				f.db,
				lateralWrap(
					quickSelectOrder("gtfs_calendar_dates", limit, nil, nil, "date").Where(In("service_id", keys)),
					"gtfs_calendars",
					"id",
					"gtfs_calendar_dates",
					"service_id",
					keys,
				),
				&ents,
			)
			return ents, err
		},
		func(ent *model.CalendarDate) int {
			return ent.ServiceID.Int()
		},
	)
}
