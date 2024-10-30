package gql

import (
	"context"

	"github.com/interline-io/transitland-lib/tt"
	"github.com/interline-io/transitland-server/model"
)

// CALENDAR

type calendarResolver struct{ *Resolver }

func (r *calendarResolver) AddedDates(ctx context.Context, obj *model.Calendar, limit *int) ([]*tt.Date, error) {
	ents, err := For(ctx).CalendarDatesByServiceID.Load(ctx, model.CalendarDateParam{ServiceID: obj.ID, Limit: checkLimit(limit), Where: nil})()
	if err != nil {
		return nil, err
	}
	ret := []*tt.Date{}
	for _, ent := range ents {
		if ent.ExceptionType.Val == 1 {
			ret = append(ret, &ent.Date)
		}
	}
	return ret, nil
}

func (r *calendarResolver) RemovedDates(ctx context.Context, obj *model.Calendar, limit *int) ([]*tt.Date, error) {
	ents, err := For(ctx).CalendarDatesByServiceID.Load(ctx, model.CalendarDateParam{ServiceID: obj.ID, Limit: checkLimit(limit), Where: nil})()
	if err != nil {
		return nil, err
	}
	ret := []*tt.Date{}
	for _, ent := range ents {
		if ent.ExceptionType.Val == 2 {
			ret = append(ret, &ent.Date)
		}
	}
	return ret, nil
}
