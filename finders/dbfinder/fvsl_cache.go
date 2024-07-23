package dbfinder

import (
	"context"
	"errors"
	"sync"
	"time"

	sq "github.com/Masterminds/squirrel"
	"github.com/interline-io/transitland-dbutil/dbutil"
	"github.com/interline-io/transitland-lib/tl"
	"github.com/interline-io/transitland-server/model"
	"github.com/jmoiron/sqlx"
)

type fvslCache struct {
	db          sqlx.Ext
	lock        sync.Mutex
	fvslWindows map[int]model.ServiceWindow
}

func newFvslCache(db sqlx.Ext) *fvslCache {
	return &fvslCache{
		db:          db,
		fvslWindows: map[int]model.ServiceWindow{},
	}
}

func (f *fvslCache) Get(ctx context.Context, fvid int) (model.ServiceWindow, bool, error) {
	f.lock.Lock()
	a, ok := f.fvslWindows[fvid]
	f.lock.Unlock()
	if ok {
		return a, ok, nil
	}
	a, err := f.query(ctx, fvid)
	f.Set(ctx, fvid, a)
	return a, true, err
}

func (f *fvslCache) Set(ctx context.Context, fvid int, w model.ServiceWindow) {
	f.lock.Lock()
	defer f.lock.Unlock()
	f.fvslWindows[fvid] = w
}

func (f *fvslCache) query(ctx context.Context, fvid int) (model.ServiceWindow, error) {
	ret := model.ServiceWindow{}
	type fvslQuery struct {
		FetchedAt    tl.Time
		StartDate    tl.Time
		EndDate      tl.Time
		TotalService tl.Int
	}
	minServiceRatio := 0.75
	startDate := time.Time{}
	endDate := time.Time{}
	bestWeek := time.Time{}

	// Get FVSLs
	q := sq.StatementBuilder.
		Select("fv.fetched_at", "fvsl.start_date", "fvsl.end_date", "monday + tuesday + wednesday + thursday + friday + saturday + sunday as total_service").
		From("feed_version_service_levels fvsl").
		Join("feed_versions fv on fv.id = fvsl.feed_version_id").
		Where(sq.Eq{"route_id": nil}).
		Where(sq.Eq{"fvsl.feed_version_id": fvid}).
		OrderBy("fvsl.start_date").
		Limit(1000)
	var ents []fvslQuery
	if err := dbutil.Select(ctx, f.db, q, &ents); err != nil {
		return ret, logErr(ctx, err)
	}
	if len(ents) == 0 {
		return ret, errors.New("no fvsl results")
	}

	var fis []tl.FeedInfo
	fiq := sq.StatementBuilder.Select("*").From("gtfs_feed_infos").Where(sq.Eq{"feed_version_id": fvid}).OrderBy("feed_start_date").Limit(1)
	if err := dbutil.Select(ctx, f.db, fiq, &fis); err != nil {
		return ret, logErr(ctx, err)
	}

	// Check if we have feed infos, otherwise calculate based on fetched week or highest service week
	fetched := ents[0].FetchedAt.Val
	if len(fis) > 0 && fis[0].FeedStartDate.Valid && fis[0].FeedEndDate.Valid {
		// fmt.Println("using feed infos")
		startDate = fis[0].FeedStartDate.Val
		endDate = fis[0].FeedEndDate.Val
	} else {
		// Get the week which includes fetched_at date, and the highest service week
		highestIdx := 0
		highestService := -1
		fetchedWeek := -1
		for i, ent := range ents {
			sd := ent.StartDate.Val
			ed := ent.EndDate.Val
			if (sd.Before(fetched) || sd.Equal(fetched)) && (ed.After(fetched) || ed.Equal(fetched)) {
				fetchedWeek = i
			}
			if ent.TotalService.Int() > highestService {
				highestIdx = i
				highestService = ent.TotalService.Int()
			}
		}
		if fetchedWeek < 0 {
			// fmt.Println("fetched week not in fvsls, using highest week:", highestIdx, highestService)
			fetchedWeek = highestIdx
		} else {
			// fmt.Println("using fetched week:", fetchedWeek)
		}
		// If the fetched week has bad service, use highest week
		if float64(ents[fetchedWeek].TotalService.Val)/float64(highestService) < minServiceRatio {
			// fmt.Println("fetched week has poor service ratio, falling back to highest week:", fetchedWeek)
			fetchedWeek = highestIdx
		}

		// Expand window in both directions from chosen week
		startDate = ents[fetchedWeek].StartDate.Val
		endDate = ents[fetchedWeek].EndDate.Val
		for i := fetchedWeek; i < len(ents); i++ {
			ent := ents[i]
			if float64(ent.TotalService.Val)/float64(highestService) < minServiceRatio {
				break
			}
			if ent.StartDate.Val.Before(startDate) {
				startDate = ent.StartDate.Val
			}
			endDate = ent.EndDate.Val
		}
		for i := fetchedWeek - 1; i > 0; i-- {
			ent := ents[i]
			if float64(ent.TotalService.Val)/float64(highestService) < minServiceRatio {
				break
			}
			if ent.EndDate.Val.After(endDate) {
				endDate = ent.EndDate.Val
			}
			startDate = ent.StartDate.Val
		}
	}

	// Check again to find the highest service week in the window
	// This will be used as the typical week for dates outside the window
	// bestWeek must start with a Monday
	bestWeek = ents[0].StartDate.Val
	bestService := ents[0].TotalService.Val
	for _, ent := range ents {
		sd := ent.StartDate.Val
		ed := ent.EndDate.Val
		if (sd.Before(endDate) || sd.Equal(endDate)) && (ed.After(startDate) || ed.Equal(startDate)) {
			if ent.TotalService.Val > bestService {
				bestService = ent.TotalService.Val
				bestWeek = ent.StartDate.Val
			}
		}
	}
	// return startDate, endDate, bestWeek, nil
	return model.ServiceWindow{
		StartDate: startDate,
		EndDate:   endDate,
		BestWeek:  bestWeek,
	}, nil
	// var err error
	// w := fvslWindow{}
	// w.StartDate, w.EndDate, w.BestWeek, err = f.finder.FindFeedVersionServiceWindow(ctx, fvid)
	// log.Trace().
	// 	Str("start_date", w.StartDate.Format("2006-01-02")).
	// 	Str("end_date", w.EndDate.Format("2006-01-02")).
	// 	Str("best_week", w.BestWeek.Format("2006-01-02")).
	// 	Int("fvid", fvid).
	// 	Msg("service window result")
	// if err != nil {
	// 	return w, err
	// }
	// w.Valid = true
	// return w, err
}
