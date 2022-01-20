package find

import (
	sq "github.com/Masterminds/squirrel"
	"github.com/interline-io/transitland-server/model"
)

func OperatorSelect(limit *int, after *int, ids []int, where *model.OperatorFilter) sq.SelectBuilder {
	qView := sq.StatementBuilder.
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
			"co.associated_feeds as associated_feeds",
		).
		From("current_operators_in_feed coif").
		Join("current_feeds on current_feeds.id = coif.feed_id").
		JoinClause("left join current_operators co on co.id = coif.operator_id").
		Where(sq.Eq{"current_feeds.deleted_at": nil}).
		Where(sq.Eq{"co.deleted_at": nil}). // not present, or present but not deleted
		OrderBy("coif.resolved_onestop_id, coif.operator_id")
	if where != nil && where.Merged != nil && *where.Merged {
		qView = qView.Distinct().Options("on (onestop_id)")
	}
	q := sq.StatementBuilder.Select("t.*").FromSelect(qView, "t")
	if len(ids) > 0 {
		q = q.Where(sq.Eq{"t.id": ids})
	}
	if after != nil {
		q = q.Where(sq.Gt{"t.id": *after})
	}
	q = q.OrderBy("id")
	q = q.Limit(checkLimit(limit))
	if where != nil {
		if where.Search != nil && len(*where.Search) > 0 {
			rank, wc := tsQuery(*where.Search)
			q = q.Column(rank).Where(wc)
		}
		if where.FeedOnestopID != nil {
			q = q.Where(sq.Eq{"feed_onestop_id": *where.FeedOnestopID})
		}
		if where.AgencyID != nil {
			q = q.Where(sq.Eq{"resolved_gtfs_agency_id": *where.AgencyID})
		}
		if where.OnestopID != nil {
			q = q.Where(sq.Eq{"onestop_id": where.OnestopID})
		}
		// Tags
		if where.Tags != nil {
			for _, k := range where.Tags.Keys() {
				if v, ok := where.Tags.Get(k); ok {
					if v == "" {
						q = q.Where("operator_tags ?? ?", k)
					} else {
						q = q.Where("operator_tags->>? = ?", k, v)
					}
				}
			}
		}
	}
	return q
}
