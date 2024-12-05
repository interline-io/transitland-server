package clock

import (
	"context"
	"sync"
	"time"

	sq "github.com/Masterminds/squirrel"
	"github.com/interline-io/transitland-dbutil/dbutil"
	"github.com/interline-io/transitland-lib/tt"
	"github.com/interline-io/transitland-mw/caches/tzcache"
	"github.com/jmoiron/sqlx"
)

type ServiceWindow struct {
	StartDate    time.Time
	EndDate      time.Time
	FallbackWeek time.Time
	Location     *time.Location
}

type ServiceWindowCache struct {
	db          sqlx.Ext
	lock        sync.Mutex
	fvslWindows map[int]ServiceWindow
	tzCache     *tzcache.Cache[int]
}

func NewServiceWindowCache(db sqlx.Ext) *ServiceWindowCache {
	return &ServiceWindowCache{
		db:          db,
		fvslWindows: map[int]ServiceWindow{},
		tzCache:     tzcache.NewCache[int](),
	}
}

func (f *ServiceWindowCache) Get(ctx context.Context, fvid int) (ServiceWindow, bool, error) {
	f.lock.Lock()
	a, ok := f.fvslWindows[fvid]
	f.lock.Unlock()
	if ok {
		return a, ok, nil
	}

	// Get timezone from FVSW data
	fvData, err := f.queryFv(ctx, fvid)
	if err != nil {
		return a, false, err
	}
	a.Location = fvData.Location

	// Get fallback week from FVSL data
	fvslData, err := f.queryFvsl(ctx, fvid)
	if err != nil {
		return a, false, err
	}
	a.FallbackWeek = tzTruncate(fvslData.FallbackWeek, a.Location)

	// Use calculated date window if not available from FVSW
	if fvData.StartDate.IsZero() || fvData.EndDate.IsZero() {
		// Use feed info date ranges if available
		a.StartDate = tzTruncate(fvslData.StartDate, a.Location)
		a.EndDate = tzTruncate(fvslData.EndDate, a.Location)
	} else {
		// Fallback to calculated date range based on FVSL data
		a.StartDate = tzTruncate(fvData.StartDate, a.Location)
		a.EndDate = tzTruncate(fvData.EndDate, a.Location)
	}

	// Add to cache
	f.fvslWindows[fvid] = a
	return a, true, err
}

// Query feed version service level records and try to find the best date.
func (f *ServiceWindowCache) queryFv(ctx context.Context, fvid int) (ServiceWindow, error) {
	ret := ServiceWindow{}
	// Query fv fetched_at and FVSW data
	type fiQuery struct {
		FetchedAt            tt.Time
		FeedStartDate        tt.Time
		FeedEndDate          tt.Time
		EarliestCalendarDate tt.Time
		LatestCalendarDate   tt.Time
		FallbackWeek         tt.Time
		DefaultTimezone      tt.String
	}
	fvq := sq.StatementBuilder.
		Select(
			"fv.fetched_at",
			"fvsw.feed_start_date",
			"fvsw.feed_end_date",
			"fvsw.earliest_calendar_date",
			"fvsw.latest_calendar_date",
			"fvsw.fallback_week",
			"fvsw.default_timezone",
		).
		From("feed_versions fv").
		LeftJoin("feed_version_service_windows fvsw on fvsw.feed_version_id = fv.id").
		Where(sq.Eq{"fvsw.feed_version_id": fvid}).
		Limit(1)
	var fis []fiQuery
	if err := dbutil.Select(ctx, f.db, fvq, &fis); err != nil {
		return ret, err
	}
	if len(fis) == 0 {
		return ret, nil
	}
	fiData := fis[0]
	if fiData.FeedStartDate.Valid && fiData.FeedEndDate.Valid {
		ret.StartDate = fiData.FeedStartDate.Val
		ret.EndDate = fiData.FeedEndDate.Val
	}
	ret.Location, _ = f.tzCache.Location(fiData.DefaultTimezone.Val)
	return ret, nil
}

func (f *ServiceWindowCache) queryFvsl(ctx context.Context, fvid int) (ServiceWindow, error) {
	ret := ServiceWindow{}
	minServiceRatio := 0.75

	// Get FVSLs
	type fvslEnt struct {
		FetchedAt    tt.Time
		StartDate    tt.Time
		EndDate      tt.Time
		TotalService tt.Int
	}
	fvslQuery := sq.StatementBuilder.
		Select(
			"fv.fetched_at",
			"fvsl.start_date",
			"fvsl.end_date",
			"monday + tuesday + wednesday + thursday + friday + saturday + sunday as total_service",
		).
		From("feed_versions fv").
		Join("feed_version_service_levels fvsl on fvsl.feed_version_id = fv.id").
		Where(sq.Eq{"route_id": nil}).
		Where(sq.Eq{"fv.id": fvid}).
		OrderBy("fvsl.start_date").
		Limit(1000)
	var fvslEnts []fvslEnt
	if err := dbutil.Select(ctx, f.db, fvslQuery, &fvslEnts); err != nil {
		return ret, err
	}
	if len(fvslEnts) == 0 {
		return ret, nil
	}

	// Get the highest service week
	highestIdx := 0
	highestService := fvslEnts[0].TotalService.Float()
	for i, ent := range fvslEnts {
		if sl := ent.TotalService.Float(); sl > highestService {
			highestIdx = i
			highestService = sl
		}
	}
	if highestService == 0 {
		return ret, nil
	}

	// Get the week containing fetched_at, defaulting to the highest service week
	selectedWeek := highestIdx
	fetchedAt := fvslEnts[0].FetchedAt.Val
	for i, ent := range fvslEnts {
		if ent.StartDate.Val.After(fetchedAt) {
			continue
		}
		if ent.EndDate.Val.Before(fetchedAt) {
			continue
		}
		if ent.TotalService.Float()/highestService < minServiceRatio {
			// fmt.Println("fetched week has poor service ratio, falling back to highest week:", i)
			continue
		}
		// fmt.Println("using fetched week:", i)
		selectedWeek = i
	}

	// Expand window in both directions from chosen week
	startDate := fvslEnts[selectedWeek].StartDate.Val
	endDate := fvslEnts[selectedWeek].EndDate.Val
	for i := selectedWeek; i < len(fvslEnts); i++ {
		ent := fvslEnts[i]
		if ent.TotalService.Float()/highestService < minServiceRatio {
			break
		}
		if ent.StartDate.Val.Before(startDate) {
			startDate = ent.StartDate.Val
		}
		endDate = ent.EndDate.Val
	}
	for i := selectedWeek - 1; i > 0; i-- {
		ent := fvslEnts[i]
		if ent.TotalService.Float()/highestService < minServiceRatio {
			break
		}
		if ent.EndDate.Val.After(endDate) {
			endDate = ent.EndDate.Val
		}
		startDate = ent.StartDate.Val
	}

	// Check again to find the highest service week in the window
	// This will be used as the typical week for dates outside the window
	// bestWeek must start with a Monday
	bestWeek := fvslEnts[0].StartDate.Val
	bestService := fvslEnts[0].TotalService.Val
	for _, ent := range fvslEnts {
		sd := ent.StartDate.Val
		ed := ent.EndDate.Val
		if (sd.Before(endDate) || sd.Equal(endDate)) && (ed.After(startDate) || ed.Equal(startDate)) {
			if ent.TotalService.Val > bestService {
				bestService = ent.TotalService.Val
				bestWeek = ent.StartDate.Val
			}
		}
	}
	return ServiceWindow{
		StartDate:    startDate,
		EndDate:      endDate,
		FallbackWeek: bestWeek,
	}, nil
}
