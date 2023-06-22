package find

import (
	sq "github.com/Masterminds/squirrel"
	"github.com/interline-io/transitland-server/model"
)

func OperatorSelect(limit *int, after *model.Cursor, ids []int, feedIds []int, userCheck *model.UserCheck, where *model.OperatorFilter) sq.SelectBuilder {
	distinct := true
	q := sq.StatementBuilder.
		Select(
			"coif.id as id",
			"coif.feed_id as feed_id",
			"coif.resolved_name as name",
			"coif.resolved_short_name as short_name",
			"coif.resolved_onestop_id as onestop_id",
			"coif.textsearch as textsearch",
			"current_feeds.onestop_id as feed_onestop_id",
			"co.file as file",
			"co.id as operator_id",
			"co.website as website",
			"co.operator_tags as operator_tags",
		).
		From("current_operators_in_feed coif").
		Join("current_feeds on current_feeds.id = coif.feed_id").
		LeftJoin("current_operators co on co.id = coif.operator_id").
		Where(sq.Eq{"current_feeds.deleted_at": nil}).
		Where(sq.Eq{"co.deleted_at": nil}). // not present, or present but not deleted
		OrderBy("coif.resolved_onestop_id, coif.operator_id")

	if where != nil {
		if where.Merged != nil && !*where.Merged {
			distinct = false
		}
		if where.FeedOnestopID != nil {
			q = q.Where(sq.Eq{"current_feeds.onestop_id": where.FeedOnestopID})
		}
		if where.AgencyID != nil {
			q = q.Where(sq.Eq{"coif.resolved_gtfs_agency_id": where.AgencyID})
		}
		if where.OnestopID != nil {
			q = q.Where(sq.Eq{"coif.resolved_onestop_id": where.OnestopID})
		}
		// Tags
		if where.Tags != nil {
			for _, k := range where.Tags.Keys() {
				if v, ok := where.Tags.Get(k); ok {
					if v == "" {
						q = q.Where("co.operator_tags ?? ?", k)
					} else {
						q = q.Where("co.operator_tags->>? = ?", k, v)
					}
				}
			}
		}
		// Places
		if where.Adm0Iso != nil || where.Adm1Iso != nil || where.Adm0Name != nil || where.Adm1Name != nil || where.CityName != nil {
			q = q.
				Join("feed_states ON feed_states.feed_id = coif.feed_id").
				Join("gtfs_agencies ON gtfs_agencies.feed_version_id = feed_states.feed_version_id AND gtfs_agencies.agency_id = coif.resolved_gtfs_agency_id").
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
	}
	if distinct {
		q = q.Distinct().Options("on (coif.resolved_onestop_id)")
	}
	if len(ids) > 0 {
		q = q.Where(sq.Eq{"coif.id": ids})
	}
	if userCheck != nil {
		q = q.Where(sq.Or{sq.Eq{"coif.feed_id": userCheck.AllowedFeeds}})
	}
	if len(feedIds) > 0 {
		q = q.Where(sq.Eq{"coif.feed_id": feedIds})
	}
	if after != nil && after.Valid && after.ID > 0 {
		q = q.Where(sq.Gt{"coif.id": after.ID})
	}

	// Outer query
	qView := sq.StatementBuilder.Select("t.*").FromSelect(q, "t").Limit(checkLimit(limit))
	if where != nil {
		if where.Search != nil && len(*where.Search) > 0 {
			rank, wc := tsQuery(*where.Search)
			qView = qView.Column(rank).Where(wc)
		}
	}
	qView = qView.OrderBy("id")
	return qView
}

func OperatorsByAgencyID(limit *int, after *model.Cursor, agencyIds []int, onestopIds []string) sq.SelectBuilder {
	q := sq.StatementBuilder.
		Select(
			"coif.id as id",
			"coif.feed_id as feed_id",
			"coif.resolved_name as name",
			"coif.resolved_short_name as short_name",
			"coif.resolved_onestop_id as onestop_id",
			"coif.textsearch as textsearch",
			"current_feeds.onestop_id as feed_onestop_id",
			"co.file as file",
			"co.id as operator_id",
			"co.website as website",
			"co.operator_tags as operator_tags",
			"a.id as agency_id",
		).
		From("current_operators_in_feed coif").
		Join("current_feeds on current_feeds.id = coif.feed_id").
		LeftJoin("current_operators co on co.id = coif.operator_id").
		Where(sq.Eq{"current_feeds.deleted_at": nil}).
		Where(sq.Eq{"co.deleted_at": nil}). // not present, or present but not deleted
		OrderBy("coif.resolved_onestop_id, coif.operator_id")

	if len(agencyIds) > 0 {
		q = q.
			Join("feed_states fs on fs.feed_id = current_feeds.id").
			Join("gtfs_agencies a on a.feed_version_id = fs.feed_version_id and a.agency_id = coif.resolved_gtfs_agency_id").
			Where(sq.Eq{"a.id": agencyIds})
	}
	if len(onestopIds) > 0 {
		q = q.Where(sq.Eq{"coif.resolved_onestop_id": onestopIds})
	}
	return q
}
