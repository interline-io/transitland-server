package resolvers

import (
	"context"

	"github.com/interline-io/transitland-server/model"
)

// AGENCY

type agencyResolver struct{ *Resolver }

func (r *agencyResolver) Cursor(ctx context.Context, obj *model.Agency) (*model.Cursor, error) {
	c := model.NewCursor(obj.FeedVersionID, obj.ID)
	return &c, nil
}

func (r *agencyResolver) Routes(ctx context.Context, obj *model.Agency, limit *int, where *model.RouteFilter) ([]*model.Route, error) {
	return For(ctx).RoutesByAgencyID.Load(ctx, model.RouteParam{AgencyID: obj.ID, Limit: limit, Where: where})()
}

func (r *agencyResolver) FeedVersion(ctx context.Context, obj *model.Agency) (*model.FeedVersion, error) {
	return For(ctx).FeedVersionsByID.Load(ctx, obj.FeedVersionID)()
}

func (r *agencyResolver) Places(ctx context.Context, obj *model.Agency, limit *int, where *model.AgencyPlaceFilter) ([]*model.AgencyPlace, error) {
	return For(ctx).AgencyPlacesByAgencyID.Load(ctx, model.AgencyPlaceParam{AgencyID: obj.ID, Limit: limit, Where: where})()
}

func (r *agencyResolver) Operator(ctx context.Context, obj *model.Agency) (*model.Operator, error) {
	if obj.CoifID == nil {
		return nil, nil
	}
	return For(ctx).OperatorsByCOIF.Load(ctx, *obj.CoifID)()

}

func (r *agencyResolver) Alerts(ctx context.Context, obj *model.Agency, active *bool, limit *int) ([]*model.Alert, error) {
	rtAlerts := r.rtfinder.FindAlertsForAgency(obj, limit, active)
	return rtAlerts, nil
}
