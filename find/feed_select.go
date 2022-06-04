package find

import (
	sq "github.com/Masterminds/squirrel"
	"github.com/interline-io/transitland-lib/log"
	"github.com/interline-io/transitland-server/model"
)

func FeedSelect(limit *int, after *model.Cursor, ids []int, where *model.FeedFilter) sq.SelectBuilder {
	q := sq.StatementBuilder.
		Select("t.*").
		From("current_feeds t").
		OrderBy("t.id asc").
		Limit(checkRange(limit, 0, 10_000)).
		Where(sq.Eq{"deleted_at": nil})
	if len(ids) > 0 {
		q = q.Where(sq.Eq{"t.id": ids})
	}
	if after != nil && after.Valid {
		q = q.Where(sq.Gt{"t.id": after.ID})
	}
	if where != nil {
		if where.Search != nil && len(*where.Search) > 1 {
			rank, wc := tsQuery(*where.Search)
			q = q.Column(rank).Where(wc)
		}
		if where.OnestopID != nil {
			q = q.Where(sq.Eq{"onestop_id": *where.OnestopID})
		}
		if len(where.Spec) > 0 {
			var specs []string
			for _, s := range where.Spec {
				specs = append(specs, s.ToDBString())
			}
			q = q.Where(sq.Eq{"spec": specs})
		}
		// Tags
		if where.Tags != nil {
			for _, k := range where.Tags.Keys() {
				if v, ok := where.Tags.Get(k); ok {
					if v == "" {
						q = q.Where("feed_tags ?? ?", k)
					} else {
						q = q.Where("feed_tags->>? = ?", k, v)
					}
				}
			}
		}
		// Fetch error
		if v := where.FetchError; v == nil {
			// nothing
		} else if *v {
			q = q.JoinClause("join lateral (select success from feed_fetches where feed_fetches.feed_id = t.id order by fetched_at desc limit 1) ff on true").Where(sq.Eq{"ff.success": false})
		} else if !*v {
			q = q.JoinClause("join lateral (select success from feed_fetches where feed_fetches.feed_id = t.id order by fetched_at desc limit 1) ff on true").Where(sq.Eq{"ff.success": true})
		}
		// Import import status
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
				log.Error().Str("value", v.String()).Msg("unknown imnport status enum")
			}
			// This lateral join gets the most recent attempt at a completed feed_version_gtfs_import and checks the status
			q = q.JoinClause(`JOIN LATERAL (select fvi.in_progress, fvi.success from feed_versions fv inner join feed_version_gtfs_imports fvi on fvi.feed_version_id = fv.id WHERE fv.feed_id = t.id ORDER BY fvi.id DESC LIMIT 1) fvicheck ON TRUE`).
				Where(sq.Eq{"fvicheck.success": checkSuccess, "fvicheck.in_progress": checkInProgress})
		}
		// Source URL
		if where.SourceURL != nil {
			urlType := "static_current"
			if where.SourceURL.Type != nil {
				urlType = where.SourceURL.Type.String()
			}
			if where.SourceURL.URL == nil {
				q = q.Where("urls->>? is not null", urlType)
			} else if v := where.SourceURL.CaseSensitive; v != nil && *v {
				q = q.Where("urls->>? = ?", urlType, where.SourceURL.URL)
			} else {
				q = q.Where("lower(urls->>?) = lower(?)", urlType, where.SourceURL.URL)
			}
		}
	}
	return q
}
