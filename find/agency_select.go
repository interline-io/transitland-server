package find

import (
	sq "github.com/Masterminds/squirrel"
	"github.com/interline-io/transitland-server/model"
	"github.com/jmoiron/sqlx"
)

func FindAgencies(atx sqlx.Ext, limit *int, after *int, ids []int, where *model.AgencyFilter) (ents []*model.Agency, err error) {
	q := AgencySelect(limit, after, ids, where)
	if len(ids) == 0 && (where == nil || where.FeedVersionSha1 == nil) {
		q = q.Where(sq.NotEq{"active": nil})
	}
	MustSelect(model.DB, q, &ents)
	return ents, nil
}

func AgencySelect(limit *int, after *int, ids []int, where *model.AgencyFilter) sq.SelectBuilder {
	qView := sq.StatementBuilder.
		Select(
			"gtfs_agencies.*",
			"tl_agency_geometries.geometry",
			"current_feeds.onestop_id AS feed_onestop_id",
			"feed_versions.sha1 AS feed_version_sha1",
			`COALESCE(
				coif.onestop_id, 
				tl_agency_onestop_ids.onestop_id::character varying, 
				(
					(('o-'::text || "right"(current_feeds.onestop_id::text, length(current_feeds.onestop_id::text) - 2)) || 
					'-'::text) || 
					regexp_replace(regexp_replace(lower(gtfs_agencies.agency_name), '[\-\:\&\@\/]', '~', 'g'), '[^[:alnum:]\~\>\<]', '', 'g')
				)::character varying
			) AS onestop_id`,
			"feed_states.feed_version_id AS active",
		).
		From("gtfs_agencies").
		Join("feed_versions ON feed_versions.id = gtfs_agencies.feed_version_id").
		Join("current_feeds ON current_feeds.id = feed_versions.feed_id").
		JoinClause("LEFT JOIN tl_agency_geometries ON tl_agency_geometries.agency_id = gtfs_agencies.id").
		JoinClause("LEFT JOIN tl_agency_onestop_ids ON tl_agency_onestop_ids.agency_id = gtfs_agencies.id").
		JoinClause("LEFT JOIN feed_states ON feed_states.feed_version_id = gtfs_agencies.feed_version_id").
		JoinClause(`LEFT JOIN (
			select co.onestop_id, coif.feed_id, gtfs_agencies.agency_id
			from current_operators co
			inner join current_operators_in_feed coif on coif.operator_id = co.id
			inner join gtfs_agencies on gtfs_agencies.id = coif.agency_id
		) coif on coif.feed_id = current_feeds.id and coif.agency_id = gtfs_agencies.agency_id`).
		Where(sq.Eq{"current_feeds.deleted_at": nil}).
		OrderBy("gtfs_agencies.id")

	q := sq.StatementBuilder.Select("*").FromSelect(qView, "t")
	if len(ids) > 0 {
		q = q.Where(sq.Eq{"t.id": ids})
	}
	if after != nil {
		q = q.Where(sq.Gt{"t.id": *after})
	}
	q = q.Limit(checkLimit(limit))
	if where != nil {
		if where.Search != nil && len(*where.Search) > 1 {
			rank, wc := tsQuery(*where.Search)
			q = q.Column(rank).Where(wc)
		}
		if where.FeedVersionSha1 != nil {
			q = q.Where(sq.Eq{"feed_version_sha1": *where.FeedVersionSha1})
		}
		if where.FeedOnestopID != nil {
			q = q.Where(sq.Eq{"feed_onestop_id": *where.FeedOnestopID})
		}
		if where.AgencyID != nil {
			q = q.Where(sq.Eq{"agency_id": *where.AgencyID})
		}
		if where.AgencyName != nil {
			q = q.Where(sq.Eq{"agency_name": *where.AgencyName})
		}
		if where.OnestopID != nil {
			q = q.Where(sq.Eq{"onestop_id": *where.OnestopID})
		}
		if where.Within != nil && where.Within.Valid {
			q = q.Where("ST_Intersects(t.geometry, ?)", where.Within)
		}
		if where.Near != nil {
			radius := checkFloat(&where.Near.Radius, 0, 10_000)
			q = q.Where("ST_DWithin(t.geometry, ST_MakePoint(?,?), ?)", where.Near.Lat, where.Near.Lon, radius)
		}
	}
	return q
}

func AgencyPlaceSelect(limit *int, after *int, ids []int, where *model.AgencyPlaceFilter) sq.SelectBuilder {
	q := quickSelectOrder("tl_agency_places", limit, after, ids, "rank desc")
	if where != nil {
		// if where.Search != nil && len(*where.Search) > 1 {
		// 	q = q.Where(tsQuery(*where.Search))
		// }
		if where.MinRank != nil {
			q = q.Where(sq.GtOrEq{"rank": where.MinRank})
		}
	}
	return q
}
