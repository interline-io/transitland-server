package dbfinder

import (
	"context"

	sq "github.com/Masterminds/squirrel"
	"github.com/interline-io/transitland-dbutil/dbutil"
	"github.com/interline-io/transitland-server/model"
)

func (f *Finder) LevelsByID(ctx context.Context, ids []int) ([]*model.Level, []error) {
	var ents []*model.Level
	err := dbutil.Select(ctx,
		f.db,
		quickSelect("gtfs_levels", nil, nil, ids),
		&ents,
	)
	if err != nil {
		return nil, logExtendErr(ctx, len(ids), err)
	}
	return arrangeBy(ids, ents, func(ent *model.Level) int { return ent.ID }), nil
}

func (f *Finder) PathwaysByID(ctx context.Context, ids []int) ([]*model.Pathway, []error) {
	var ents []*model.Pathway
	err := dbutil.Select(ctx,
		f.db,
		pathwaySelect(nil, nil, ids, nil, nil),
		&ents,
	)
	if err != nil {
		return nil, logExtendErr(ctx, len(ids), err)
	}
	return arrangeBy(ids, ents, func(ent *model.Pathway) int { return ent.ID }), nil
}

func (f *Finder) LevelsByParentStationID(ctx context.Context, params []model.LevelParam) ([][]*model.Level, []error) {
	return paramGroupQuery(
		params,
		func(p model.LevelParam) (int, bool, *int) {
			return p.ParentStationID, false, p.Limit
		},
		func(keys []int, where bool, limit *int) (ents []*model.Level, err error) {
			err = dbutil.Select(ctx,
				f.db,
				lateralWrap(
					quickSelect("gtfs_levels", limit, nil, nil),
					"gtfs_stops",
					"id",
					"gtfs_levels",
					"parent_station",
					keys,
				),
				&ents,
			)
			return ents, err
		},
		func(ent *model.Level) int {
			return ent.ParentStation.Int()
		},
	)
}

func (f *Finder) PathwaysByFromStopID(ctx context.Context, params []model.PathwayParam) ([][]*model.Pathway, []error) {
	return paramGroupQuery(
		params,
		func(p model.PathwayParam) (int, *model.PathwayFilter, *int) {
			return p.FromStopID, p.Where, p.Limit
		},
		func(keys []int, where *model.PathwayFilter, limit *int) (ents []*model.Pathway, err error) {
			err = dbutil.Select(ctx,
				f.db,
				lateralWrap(
					pathwaySelect(limit, nil, nil, f.PermFilter(ctx), where),
					"gtfs_stops",
					"id",
					"gtfs_pathways",
					"from_stop_id",
					keys,
				),
				&ents,
			)
			return ents, err
		},
		func(ent *model.Pathway) int {
			return ent.FromStopID.Int()
		},
	)
}

func (f *Finder) PathwaysByToStopID(ctx context.Context, params []model.PathwayParam) ([][]*model.Pathway, []error) {
	return paramGroupQuery(
		params,
		func(p model.PathwayParam) (int, *model.PathwayFilter, *int) {
			return p.ToStopID, p.Where, p.Limit
		},
		func(keys []int, where *model.PathwayFilter, limit *int) (ents []*model.Pathway, err error) {
			err = dbutil.Select(ctx,
				f.db,
				lateralWrap(
					pathwaySelect(limit, nil, nil, f.PermFilter(ctx), where),
					"gtfs_stops",
					"id",
					"gtfs_pathways",
					"to_stop_id",
					keys,
				),
				&ents,
			)
			return ents, err
		},
		func(ent *model.Pathway) int {
			return ent.ToStopID.Int()
		},
	)
}

func pathwaySelect(limit *int, after *model.Cursor, ids []int, permFilter *model.PermFilter, where *model.PathwayFilter) sq.SelectBuilder {
	q := sq.StatementBuilder.
		Select(
			"gtfs_pathways.id",
			"gtfs_pathways.feed_version_id",
			"gtfs_pathways.pathway_id",
			"gtfs_pathways.from_stop_id",
			"gtfs_pathways.to_stop_id",
			"gtfs_pathways.pathway_mode",
			"gtfs_pathways.is_bidirectional",
			"gtfs_pathways.length",
			"gtfs_pathways.traversal_time",
			"gtfs_pathways.stair_count",
			"gtfs_pathways.max_slope",
			"gtfs_pathways.min_width",
			"gtfs_pathways.signposted_as",
			"gtfs_pathways.reverse_signposted_as",
		).
		From("gtfs_pathways").
		Join("feed_versions on feed_versions.id = gtfs_pathways.feed_version_id").
		Join("current_feeds on current_feeds.id = feed_versions.feed_id").
		Limit(checkLimit(limit)).
		OrderBy("gtfs_pathways.id")

	if where != nil {
		if where.PathwayMode != nil {
			q = q.Where(sq.Eq{"pathway_mode": where.PathwayMode})
		}
	}
	if len(ids) > 0 {
		q = q.Where(In("gtfs_pathways.id", ids))
	}
	if after != nil && after.Valid && after.ID > 0 {
		q = q.Where(sq.Gt{"gtfs_pathways.id": after.ID})
	}

	// Handle permissions
	q = pfJoinCheckFv(q, permFilter)
	return q
}
