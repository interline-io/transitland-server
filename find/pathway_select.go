package find

import (
	sq "github.com/Masterminds/squirrel"
	"github.com/interline-io/transitland-server/model"
)

func PathwaySelect(limit *int, after *model.Cursor, ids []int, userCheck *model.UserCheck, where *model.PathwayFilter) sq.SelectBuilder {
	q := quickSelectOrder("gtfs_pathways", limit, after, ids, "")
	if where != nil {
		if where.PathwayMode != nil {
			q = q.Where(sq.Eq{"pathway_mode": where.PathwayMode})
		}
	}
	return q
}
