package dbfinder

import (
	"context"

	sq "github.com/Masterminds/squirrel"
	"github.com/interline-io/transitland-dbutil/dbutil"
	"github.com/interline-io/transitland-server/model"
)

func (f *Finder) SegmentsByFeedVersionID(ctx context.Context, params []model.SegmentParam) ([][]*model.Segment, []error) {
	return paramGroupQuery(
		params,
		func(p model.SegmentParam) (int, string, *int) {
			return p.FeedVersionID, p.Layer, p.Limit
		},
		func(keys []int, layer string, limit *int) (ents []*model.Segment, err error) {
			err = dbutil.Select(ctx,
				f.db,
				lateralWrap(
					quickSelect("tl_segments", limit, nil, nil),
					"feed_versions",
					"id",
					"tl_segments",
					"feed_version_id",
					keys,
				),
				&ents,
			)
			return ents, err
		},
		func(ent *model.Segment) int {
			return ent.FeedVersionID
		},
	)
}

func (f *Finder) SegmentsByID(ctx context.Context, ids []int) ([]*model.Segment, []error) {
	var ents []*model.Segment
	err := dbutil.Select(ctx,
		f.db,
		quickSelect("tl_segments", nil, nil, ids),
		&ents,
	)
	if err != nil {
		return nil, logExtendErr(ctx, len(ids), err)
	}
	return arrangeBy(ids, ents, func(ent *model.Segment) int { return ent.ID }), nil
}

func (f *Finder) SegmentsByRouteID(ctx context.Context, params []model.SegmentParam) ([][]*model.Segment, []error) {
	type qent struct {
		RouteID int
		model.Segment
	}
	qentGroups, err := paramGroupQuery(
		params,
		func(p model.SegmentParam) (int, *model.SegmentFilter, *int) {
			return p.RouteID, p.Where, p.Limit
		},
		func(keys []int, where *model.SegmentFilter, limit *int) (ents []*qent, err error) {
			q := sq.Select("s.id", "s.way_id", "s.geometry", "s.route_id").
				From("gtfs_routes").
				JoinClause(
					`join lateral (select distinct on (tl_segments.id, tl_segment_patterns.route_id) tl_segments.id, tl_segments.way_id, tl_segments.geometry, tl_segment_patterns.route_id from tl_segments join tl_segment_patterns on tl_segment_patterns.segment_id = tl_segments.id where tl_segment_patterns.route_id = gtfs_routes.id limit ?) s on true`,
					checkLimit(limit),
				).
				Where(In("gtfs_routes.id", keys))
			err = dbutil.Select(ctx,
				f.db,
				q,
				&ents,
			)
			return ents, err
		},
		func(ent *qent) int {
			return ent.RouteID
		},
	)
	return convertEnts(qentGroups, func(a *qent) *model.Segment { return &a.Segment }), err
}

func (f *Finder) SegmentPatternsByRouteID(ctx context.Context, params []model.SegmentPatternParam) ([][]*model.SegmentPattern, []error) {
	return paramGroupQuery(
		params,
		func(p model.SegmentPatternParam) (int, *model.SegmentPatternFilter, *int) {
			return p.RouteID, p.Where, p.Limit
		},
		func(keys []int, where *model.SegmentPatternFilter, limit *int) (ents []*model.SegmentPattern, err error) {
			err = dbutil.Select(ctx,
				f.db,
				lateralWrap(
					quickSelect("tl_segment_patterns", limit, nil, nil),
					"gtfs_routes",
					"id",
					"tl_segment_patterns",
					"route_id",
					keys,
				),
				&ents,
			)
			return ents, err
		},
		func(ent *model.SegmentPattern) int {
			return ent.RouteID
		},
	)
}

func (f *Finder) SegmentPatternsBySegmentID(ctx context.Context, params []model.SegmentPatternParam) ([][]*model.SegmentPattern, []error) {
	return paramGroupQuery(
		params,
		func(p model.SegmentPatternParam) (int, *model.SegmentPatternFilter, *int) {
			return p.SegmentID, p.Where, p.Limit
		},
		func(keys []int, where *model.SegmentPatternFilter, limit *int) (ents []*model.SegmentPattern, err error) {
			err = dbutil.Select(ctx,
				f.db,
				lateralWrap(
					quickSelect("tl_segment_patterns", limit, nil, nil),
					"tl_segments",
					"id",
					"tl_segment_patterns",
					"segment_id",
					keys,
				),
				&ents,
			)
			return ents, err
		},
		func(ent *model.SegmentPattern) int {
			return ent.SegmentID
		},
	)
}
