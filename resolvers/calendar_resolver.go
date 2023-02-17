package resolvers

import (
	"context"

	"github.com/interline-io/transitland-lib/tl"
	"github.com/interline-io/transitland-lib/tl/tt"
	"github.com/interline-io/transitland-server/model"
)

// CALENDAR

type calendarResolver struct{ *Resolver }

// StartDate map time.Time to tl.Date
func (r *calendarResolver) StartDate(ctx context.Context, obj *model.Calendar) (*tl.Date, error) {
	a := tt.NewDate(obj.StartDate)
	return &a, nil
}

// EndDate map time.Time to tl.Date
func (r *calendarResolver) EndDate(ctx context.Context, obj *model.Calendar) (*tl.Date, error) {
	a := tt.NewDate(obj.EndDate)
	return &a, nil
}

func (r *calendarResolver) AddedDates(ctx context.Context, obj *model.Calendar, limit *int) ([]*tl.Date, error) {
	ents, err := For(ctx).CalendarDatesByServiceID.Load(ctx, model.CalendarDateParam{ServiceID: obj.ID, Limit: limit, Where: nil})()
	if err != nil {
		return nil, err
	}
	ret := []*tl.Date{}
	for _, ent := range ents {
		if ent.ExceptionType == 1 {
			x := tt.NewDate(ent.Date)
			ret = append(ret, &x)
		}
	}
	return ret, nil
}

func (r *calendarResolver) RemovedDates(ctx context.Context, obj *model.Calendar, limit *int) ([]*tl.Date, error) {
	ents, err := For(ctx).CalendarDatesByServiceID.Load(ctx, model.CalendarDateParam{ServiceID: obj.ID, Limit: limit, Where: nil})()
	if err != nil {
		return nil, err
	}
	ret := []*tl.Date{}
	for _, ent := range ents {
		if ent.ExceptionType == 2 {
			x := tt.NewDate(ent.Date)
			ret = append(ret, &x)
		}
	}
	return ret, nil
}
