package find

import (
	sq "github.com/Masterminds/squirrel"
	"github.com/interline-io/transitland-server/model"
	"github.com/jmoiron/sqlx"
)

func FindOperators(atx sqlx.Ext, limit *int, after *int, ids []int, where *model.OperatorFilter) (ents []*model.Operator, err error) {
	q := OperatorSelect(limit, after, ids, where)
	MustSelect(model.DB, q, &ents)
	return ents, nil
}

func OperatorSelect(limit *int, after *int, ids []int, where *model.OperatorFilter) sq.SelectBuilder {
	qView := sq.StatementBuilder.
		Select(
			"coalesce(co.id, a2.id + 100000000) as id",
			"coalesce(co.onestop_id, a2.onestop_id, a2.backup_onestop_id) as onestop_id",
			"current_feeds.onestop_id as feed_onestop_id",
			"a2.id as agency_id",
			"a2.feed_version_sha1 as feed_version_sha1",
			"co.id as operator_id",
			"co.name as operator_name",
			"co.short_name as operator_short_name",
			"co.website as operator_website",
			"co.operator_tags as operator_tags",
			"co.associated_feeds as operator_associated_feeds",
			"co.textsearch as textsearch",
		).
		From("current_operators co").
		Join("current_operators_in_feed coif on coif.operator_id = co.id").
		Join("current_feeds on current_feeds.id = coif.feed_id").
		JoinClause("left join gtfs_agencies a1 on a1.id = coif.agency_id").
		JoinClause(`full outer join (
			select 
				a2.id, 
				a2.agency_id, 
				a2.agency_name,
				feed_states.feed_id,
				fv.sha1 as feed_version_sha1,			
				tlao.onestop_id,
				(
					('o-'::text || "right"(cf.onestop_id::text, length(cf.onestop_id::text) - 2)) || 
					('-'::text) || 
					regexp_replace(regexp_replace(lower(a2.agency_name), '[\-\:\&\@\/]', '~', 'g'), '[^[:alnum:]\~\>\<]', '', 'g')
				) as backup_onestop_id
			from gtfs_agencies a2 
			inner join feed_states on feed_states.feed_version_id = a2.feed_version_id			
			inner join feed_versions fv on fv.id = feed_states.feed_version_id
			inner join current_feeds cf on cf.id = feed_states.feed_id
			left join tl_agency_onestop_ids tlao on tlao.agency_id = a2.id    
			) a2 on a2.agency_id = a1.agency_id and a2.feed_id = coif.feed_id`).
		Where(sq.Eq{"current_feeds.deleted_at": nil}).
		Where(sq.Eq{"co.deleted_at": nil}) // not present, or present but not deleted

	if where != nil && where.Merged != nil && *where.Merged {
		qView = qView.Distinct().Options("on (onestop_id)")
	}

	q := sq.StatementBuilder.Select("*").FromSelect(qView, "t")
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
		if where.FeedVersionSha1 != nil {
			q = q.Where(sq.Eq{"feed_version_sha1": *where.FeedVersionSha1})
		}
		if where.FeedOnestopID != nil {
			q = q.Where(sq.Eq{"feed_onestop_id": *where.FeedOnestopID})
		}
		if where.AgencyID != nil {
			q = q.Where(sq.Eq{"agency_id": *where.AgencyID})
		}
		if where.OnestopID != nil {
			q = q.Where(sq.Eq{"onestop_id": where.OnestopID})
		}
	}
	return q
}
