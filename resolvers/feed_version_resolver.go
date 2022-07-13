package resolvers

import (
	"context"

	"github.com/interline-io/transitland-server/model"
)

// FEED VERSION

type feedVersionResolver struct{ *Resolver }

func (r *feedVersionResolver) Cursor(ctx context.Context, obj *model.FeedVersion) (*model.Cursor, error) {
	c := model.NewCursor(0, obj.ID)
	return &c, nil
}

func (r *feedVersionResolver) Agencies(ctx context.Context, obj *model.FeedVersion, limit *int, where *model.AgencyFilter) ([]*model.Agency, error) {
	return For(ctx).AgenciesByFeedVersionID.Load(model.AgencyParam{FeedVersionID: obj.ID, Limit: limit})
}

func (r *feedVersionResolver) Routes(ctx context.Context, obj *model.FeedVersion, limit *int, where *model.RouteFilter) ([]*model.Route, error) {
	return For(ctx).RoutesByFeedVersionID.Load(model.RouteParam{FeedVersionID: obj.ID, Limit: limit, Where: where})
}

func (r *feedVersionResolver) Stops(ctx context.Context, obj *model.FeedVersion, limit *int, where *model.StopFilter) ([]*model.Stop, error) {
	return For(ctx).StopsByFeedVersionID.Load(model.StopParam{FeedVersionID: obj.ID, Limit: limit, Where: where})
}

func (r *feedVersionResolver) Trips(ctx context.Context, obj *model.FeedVersion, limit *int, where *model.TripFilter) ([]*model.Trip, error) {
	return For(ctx).TripsByFeedVersionID.Load(model.TripParam{FeedVersionID: obj.ID, Limit: limit, Where: where})
}

func (r *feedVersionResolver) Feed(ctx context.Context, obj *model.FeedVersion) (*model.Feed, error) {
	return For(ctx).FeedsByID.Load(obj.FeedID)
}

func (r *feedVersionResolver) Files(ctx context.Context, obj *model.FeedVersion, limit *int) ([]*model.FeedVersionFileInfo, error) {
	return For(ctx).FeedVersionFileInfosByFeedVersionID.Load(model.FeedVersionFileInfoParam{FeedVersionID: obj.ID, Limit: limit})
}

func (r *feedVersionResolver) FeedVersionGtfsImport(ctx context.Context, obj *model.FeedVersion) (*model.FeedVersionGtfsImport, error) {
	return For(ctx).FeedVersionGtfsImportsByFeedVersionID.Load(obj.ID)
}

func (r *feedVersionResolver) ServiceLevels(ctx context.Context, obj *model.FeedVersion, limit *int, where *model.FeedVersionServiceLevelFilter) ([]*model.FeedVersionServiceLevel, error) {
	return For(ctx).FeedVersionServiceLevelsByFeedVersionID.Load(model.FeedVersionServiceLevelParam{FeedVersionID: obj.ID, Limit: limit, Where: where})
}

func (r *feedVersionResolver) FeedInfos(ctx context.Context, obj *model.FeedVersion, limit *int) ([]*model.FeedInfo, error) {
	return For(ctx).FeedInfosByFeedVersionID.Load(model.FeedInfoParam{FeedVersionID: obj.ID, Limit: limit})
}

// FEED VERSION GTFS IMPORT

type feedVersionGtfsImportResolver struct{ *Resolver }

func (r *feedVersionGtfsImportResolver) EntityCount(ctx context.Context, obj *model.FeedVersionGtfsImport) (interface{}, error) {
	return obj.EntityCount, nil
}

func (r *feedVersionGtfsImportResolver) WarningCount(ctx context.Context, obj *model.FeedVersionGtfsImport) (interface{}, error) {
	return obj.WarningCount, nil
}

func (r *feedVersionGtfsImportResolver) SkipEntityErrorCount(ctx context.Context, obj *model.FeedVersionGtfsImport) (interface{}, error) {
	return obj.SkipEntityErrorCount, nil
}

func (r *feedVersionGtfsImportResolver) SkipEntityReferenceCount(ctx context.Context, obj *model.FeedVersionGtfsImport) (interface{}, error) {
	return obj.SkipEntityReferenceCount, nil
}

func (r *feedVersionGtfsImportResolver) SkipEntityFilterCount(ctx context.Context, obj *model.FeedVersionGtfsImport) (interface{}, error) {
	return obj.SkipEntityFilterCount, nil
}

func (r *feedVersionGtfsImportResolver) SkipEntityMarkedCount(ctx context.Context, obj *model.FeedVersionGtfsImport) (interface{}, error) {
	return obj.SkipEntityMarkedCount, nil
}

func (r *feedStateResolver) FeedVersion(ctx context.Context, obj *model.FeedState) (*model.FeedVersion, error) {
	return For(ctx).FeedVersionsByID.Load(int(obj.FeedVersionID.Val))
}
