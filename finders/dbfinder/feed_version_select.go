package dbfinder

import (
	sq "github.com/Masterminds/squirrel"
	"github.com/interline-io/log"
	"github.com/interline-io/transitland-lib/tl/tt"
	"github.com/interline-io/transitland-server/model"
)

func FeedVersionSelect(limit *int, after *model.Cursor, ids []int, permFilter *model.PermFilter, where *model.FeedVersionFilter) sq.SelectBuilder {
	q := sq.StatementBuilder.
		Select(
			"feed_versions.id",
			"feed_versions.feed_id",
			"feed_versions.sha1",
			"feed_versions.fetched_at",
			"feed_versions.url",
			"feed_versions.earliest_calendar_date",
			"feed_versions.latest_calendar_date",
			"feed_versions.created_by",
			"feed_versions.updated_by",
			"feed_versions.name",
			"feed_versions.description",
			"feed_versions.file",
		).
		From("feed_versions").
		Join("current_feeds on current_feeds.id = feed_versions.feed_id").Where(sq.Eq{"current_feeds.deleted_at": nil}).
		Limit(checkLimit(limit)).
		OrderBy("feed_versions.fetched_at desc, feed_versions.id desc")

	if where != nil {
		if where.Sha1 != nil {
			q = q.Where(sq.Eq{"feed_versions.sha1": *where.Sha1})
		}
		if where.File != nil {
			q = q.Where(sq.Eq{"feed_versions.file": where.File})
		}
		if len(where.FeedIds) > 0 {
			q = q.Where(sq.Eq{"feed_versions.feed_id": where.FeedIds})
		}
		if where.FeedOnestopID != nil {
			q = q.Where(sq.Eq{"current_feeds.onestop_id": *where.FeedOnestopID})
		}

		// Spatial
		if where.Bbox != nil || where.Within != nil || where.Near != nil {
			q = q.Join("tl_feed_version_geometries fv_geoms on fv_geoms.feed_version_id = feed_versions.id")
			if where.Bbox != nil {
				q = q.Where("ST_Intersects(fv_geoms.geometry, ST_MakeEnvelope(?,?,?,?,4326))", where.Bbox.MinLon, where.Bbox.MinLat, where.Bbox.MaxLon, where.Bbox.MaxLat)
			}
			if where.Within != nil && where.Within.Valid {
				q = q.Where("ST_Intersects(fv_geoms.geometry, ?)", where.Within)
			}
			if where.Near != nil {
				radius := checkFloat(&where.Near.Radius, 0, 1_000_000)
				q = q.Where("ST_DWithin(fv_geoms.geometry, ST_MakePoint(?,?), ?)", where.Near.Lon, where.Near.Lat, radius)
			}
		}

		// Coverage
		if covers := where.Covers; covers != nil {
			joinFvsw := false
			if covers.StartDate != nil && covers.StartDate.Valid {
				joinFvsw = true
				q = q.
					Where(sq.LtOrEq{"coalesce(fvsw.feed_start_date,fvsw.earliest_calendar_date)": covers.StartDate.Val}).
					Where(sq.GtOrEq{"coalesce(fvsw.feed_end_date,fvsw.latest_calendar_date)": covers.StartDate.Val})
			}
			if covers.EndDate != nil && covers.EndDate.Valid {
				joinFvsw = true
				q = q.
					Where(sq.LtOrEq{"coalesce(fvsw.feed_start_date,fvsw.earliest_calendar_date)": covers.EndDate.Val}).
					Where(sq.GtOrEq{"coalesce(fvsw.feed_end_date,fvsw.latest_calendar_date)": covers.EndDate.Val})
			}
			if joinFvsw {
				q = q.Join("feed_version_service_windows fvsw on fvsw.feed_version_id = feed_versions.id")
			}
			if covers.FetchedBefore != nil {
				q = q.Where(sq.Lt{"feed_versions.fetched_at": covers.FetchedBefore})
			}
			if covers.FetchedAfter != nil {
				q = q.Where(sq.Gt{"feed_versions.fetched_at": covers.FetchedAfter})
			}
		}

		// Import import status
		// Similar logic to FeedSelect
		if where.ImportStatus != nil {
			// in_progress must be false to check success and vice-versa
			var checkSuccess bool
			var checkInProgress bool
			switch v := *where.ImportStatus; v {
			case model.ImportStatusSuccess:
				checkSuccess = true
				checkInProgress = false
			case model.ImportStatusInProgress:
				checkSuccess = false
				checkInProgress = true
			case model.ImportStatusError:
				checkSuccess = false
				checkInProgress = false
			default:
				log.Error().Str("value", v.String()).Msg("unknown import status enum")
			}
			q = q.Join(`feed_version_gtfs_imports fvgi on fvgi.feed_version_id = feed_versions.id`).
				Where(sq.Eq{"fvgi.success": checkSuccess, "fvgi.in_progress": checkInProgress})
		}
	}
	if len(ids) > 0 {
		q = q.Where(sq.Eq{"feed_versions.id": ids})
	}
	if after != nil && after.Valid && after.ID > 0 {
		q = q.Where(sq.Expr("(feed_versions.fetched_at,feed_versions.id) < (select fetched_at,id from feed_versions where id = ?)", after.ID))
	}

	// Handle permissions
	q = q.
		Join("feed_states fsp on fsp.feed_id = current_feeds.id").
		Where(sq.Or{
			sq.Expr("fsp.public = true"),
			sq.Eq{"fsp.feed_id": permFilter.GetAllowedFeeds()},
			sq.Eq{"feed_versions.id": permFilter.GetAllowedFeedVersions()},
		})
	return q
}

func FeedVersionServiceLevelSelect(limit *int, after *model.Cursor, ids []int, permFilter *model.PermFilter, where *model.FeedVersionServiceLevelFilter) sq.SelectBuilder {
	q := sq.StatementBuilder.
		Select(
			"feed_version_service_levels.id",
			"feed_version_service_levels.feed_version_id",
			"feed_version_service_levels.route_id",
			"feed_version_service_levels.start_date",
			"feed_version_service_levels.end_date",
			"feed_version_service_levels.monday",
			"feed_version_service_levels.tuesday",
			"feed_version_service_levels.wednesday",
			"feed_version_service_levels.thursday",
			"feed_version_service_levels.friday",
			"feed_version_service_levels.saturday",
			"feed_version_service_levels.sunday",
		).
		From("feed_version_service_levels").
		Limit(checkLimit(limit)).
		OrderBy("feed_version_service_levels.id")

	if where == nil {
		where = &model.FeedVersionServiceLevelFilter{}
	}
	q = q.Where(sq.Eq{"route_id": nil})
	if where.StartDate != nil {
		q = q.Where(sq.GtOrEq{"start_date": where.StartDate})
	}
	if where.EndDate != nil {
		q = q.Where(sq.LtOrEq{"end_date": where.EndDate})
	}
	if len(ids) > 0 {
		q = q.Where(sq.Eq{"feed_version_service_levels.id": ids})
	}
	if after != nil && after.Valid && after.ID > 0 {
		q = q.Where(sq.Gt{"feed_version_service_levels.id": after.ID})
	}
	return q
}

type FeedVersionGeometry struct {
	FeedVersionID int
	Geometry      *tt.Polygon
}

func FeedVersionGeometrySelect(ids []int) sq.SelectBuilder {
	return sq.StatementBuilder.Select("feed_version_id", "geometry").From("tl_feed_version_geometries").Where(sq.Eq{"feed_version_id": ids})
}
