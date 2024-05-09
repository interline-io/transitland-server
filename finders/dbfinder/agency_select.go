package dbfinder

import (
	sq "github.com/Masterminds/squirrel"
	"github.com/interline-io/transitland-server/model"
)

func AgencySelect(limit *int, after *model.Cursor, ids []int, active bool, permFilter *model.PermFilter, where *model.AgencyFilter) sq.SelectBuilder {
	distinct := false
	q := sq.StatementBuilder.
		Select(
			"gtfs_agencies.id",
			"gtfs_agencies.feed_version_id",
			"gtfs_agencies.agency_id",
			"gtfs_agencies.agency_name",
			"gtfs_agencies.agency_url",
			"gtfs_agencies.agency_timezone",
			"gtfs_agencies.agency_lang",
			"gtfs_agencies.agency_phone",
			"gtfs_agencies.agency_fare_url",
			"gtfs_agencies.agency_email",
			"tl_agency_geometries.geometry",
			"feed_versions.sha1 AS feed_version_sha1",
			"current_feeds.onestop_id AS feed_onestop_id",
			"coalesce (coif.resolved_onestop_id, '') as onestop_id",
			"coif.id as coif_id",
		).
		From("gtfs_agencies").
		Join("feed_versions ON feed_versions.id = gtfs_agencies.feed_version_id").
		Join("current_feeds ON current_feeds.id = feed_versions.feed_id").
		JoinClause("left join tl_agency_geometries ON tl_agency_geometries.agency_id = gtfs_agencies.id").
		JoinClause("left join current_operators_in_feed coif ON coif.feed_id = current_feeds.id AND coif.resolved_gtfs_agency_id = gtfs_agencies.agency_id").
		Where(sq.Eq{"current_feeds.deleted_at": nil}).
		OrderBy("gtfs_agencies.feed_version_id,gtfs_agencies.id").
		Limit(checkLimit(limit))

	if where != nil {
		if where.FeedVersionSha1 != nil {
			q = q.Where("feed_versions.id = (select id from feed_versions where sha1 = ? limit 1)", *where.FeedVersionSha1)
		}
		if where.FeedOnestopID != nil {
			q = q.Where(sq.Eq{"current_feeds.onestop_id": *where.FeedOnestopID})
		}
		if where.AgencyID != nil {
			q = q.Where(sq.Eq{"gtfs_agencies.agency_id": *where.AgencyID})
		}
		if where.AgencyName != nil {
			q = q.Where(sq.Eq{"gtfs_agencies.agency_name": *where.AgencyName})
		}
		if where.OnestopID != nil {
			q = q.Where(sq.Eq{"coif.resolved_onestop_id": *where.OnestopID})
		}
		// Spatial
		if where.Bbox != nil {
			q = q.Where("ST_Intersects(tl_agency_geometries.geometry, ST_MakeEnvelope(?,?,?,?,4326))", where.Bbox.MinLon, where.Bbox.MinLat, where.Bbox.MaxLon, where.Bbox.MaxLat)
		}
		if where.Within != nil && where.Within.Valid {
			q = q.Where("ST_Intersects(tl_agency_geometries.geometry, ?)", where.Within)
		}
		if where.Near != nil {
			radius := checkFloat(&where.Near.Radius, 0, 1_000_000)
			q = q.Where("ST_DWithin(tl_agency_geometries.geometry, ST_MakePoint(?,?), ?)", where.Near.Lon, where.Near.Lat, radius)
		}
		// Places
		if where.Adm0Iso != nil || where.Adm1Iso != nil || where.Adm0Name != nil || where.Adm1Name != nil || where.CityName != nil {
			distinct = true
			q = q.
				Join("tl_agency_places tlap ON tlap.agency_id = gtfs_agencies.id").
				Join("ne_10m_admin_1_states_provinces ne_admin on ne_admin.name = tlap.adm1name and ne_admin.admin = tlap.adm0name")
			if where.Adm0Iso != nil {
				q = q.Where(sq.ILike{"ne_admin.iso_a2": *where.Adm0Iso})
			}
			if where.Adm1Iso != nil {
				q = q.Where(sq.ILike{"ne_admin.iso_3166_2": *where.Adm1Iso})
			}
			if where.Adm0Name != nil {
				q = q.Where(sq.ILike{"tlap.adm0name": *where.Adm0Name})
			}
			if where.Adm1Name != nil {
				q = q.Where(sq.ILike{"tlap.adm1name": *where.Adm1Name})
			}
			if where.CityName != nil {
				q = q.Where(sq.ILike{"tlap.name": *where.CityName})
			}
		}
		// Handle license filtering
		q = licenseFilter(where.License, q)

		// Text search
		if where.Search != nil && len(*where.Search) > 1 {
			rank, wc := tsTableQuery("gtfs_agencies", *where.Search)
			q = q.Column(rank).Where(wc)
		}
	}

	if distinct {
		q = q.Distinct().Options("on (gtfs_agencies.feed_version_id,gtfs_agencies.id)")
	}
	if len(ids) > 0 {
		q = q.Where(sq.Eq{"gtfs_agencies.id": ids})
	}
	if active {
		q = q.Join("feed_states on feed_states.feed_version_id = gtfs_agencies.feed_version_id")
	}

	// Handle cursor
	if after != nil && after.Valid && after.ID > 0 {
		// first check helps improve query performance
		if after.FeedVersionID == 0 {
			q = q.
				Where(sq.Expr("gtfs_agencies.feed_version_id >= (select feed_version_id from gtfs_agencies where id = ?)", after.ID)).
				Where(sq.Expr("(gtfs_agencies.feed_version_id, gtfs_agencies.id) > (select feed_version_id,id from gtfs_agencies where id = ?)", after.ID))
		} else {
			q = q.
				Where(sq.Expr("gtfs_agencies.feed_version_id >= ?", after.FeedVersionID)).
				Where(sq.Expr("(gtfs_agencies.feed_version_id, gtfs_agencies.id) > (?,?)", after.FeedVersionID, after.ID))
		}
	}

	// Handle permissions
	q = q.
		Join("feed_states fsp on fsp.feed_id = current_feeds.id").
		Where(sq.Or{
			sq.Expr("fsp.public = true"),
			sq.Eq{"true": permFilter.IsGlobalAdmin()},
			sq.Eq{"fsp.feed_id": permFilter.GetAllowedFeeds()},
			sq.Eq{"feed_versions.id": permFilter.GetAllowedFeedVersions()},
		})

	return q
}
