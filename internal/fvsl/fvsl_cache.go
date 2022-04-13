package fvsl

import (
	"errors"
	"fmt"
	"sync"
	"time"

	sq "github.com/Masterminds/squirrel"

	"github.com/interline-io/transitland-lib/tl"
	"github.com/interline-io/transitland-server/find"
	"github.com/interline-io/transitland-server/model"
)

type FVSLWindow struct {
	FetchedAt time.Time
	StartDate time.Time
	EndDate   time.Time
	BestWeek  time.Time
	Valid     bool
}

type FVSLCache struct {
	Finder model.Finder
	lock   sync.Mutex
	fvids  map[int]FVSLWindow
}

func (f *FVSLCache) Get(fvid int) (FVSLWindow, bool) {
	f.lock.Lock()
	a, ok := f.fvids[fvid]
	f.lock.Unlock()
	if ok {
		return a, ok
	}
	a, err := f.query(fvid)
	if err != nil {
		a.Valid = false
	}
	f.Set(fvid, a)
	return a, a.Valid
}

func (f *FVSLCache) Set(fvid int, w FVSLWindow) {
	f.lock.Lock()
	defer f.lock.Unlock()
	if f.fvids == nil {
		f.fvids = map[int]FVSLWindow{}
	}
	f.fvids[fvid] = w
}

func (f *FVSLCache) query(fvid int) (FVSLWindow, error) {
	var err error
	w := FVSLWindow{}
	w.StartDate, w.EndDate, w.BestWeek, err = FindFeedVersionServiceWindow(f.Finder, fvid)
	if err != nil {
		return w, err
	}
	w.Valid = true
	return w, err
}

type fvslQuery struct {
	FetchedAt    tl.Time
	StartDate    tl.Time
	EndDate      tl.Time
	TotalService tl.Int
}

func FindFeedVersionServiceWindow(finder model.Finder, fvid int) (time.Time, time.Time, time.Time, error) {
	minServiceRatio := 0.75
	startDate := time.Time{}
	endDate := time.Time{}
	bestWeek := time.Time{}

	// Queries
	q := sq.StatementBuilder.
		Select("fv.fetched_at", "fvsl.start_date", "fvsl.end_date", "monday + tuesday + wednesday + thursday + friday + saturday + sunday as total_service").
		From("feed_version_service_levels fvsl").
		Join("feed_versions fv on fv.id = fvsl.feed_version_id").
		Where(sq.Eq{"route_id": nil}).
		Where(sq.Eq{"fvsl.feed_version_id": fvid}).
		OrderBy("fvsl.start_date").
		Limit(1000)
	var ents []fvslQuery
	find.MustSelect(finder.DBX(), q, &ents)
	if len(ents) == 0 {
		return startDate, endDate, bestWeek, errors.New("no fvsl results")
	}
	var fis []tl.FeedInfo
	fiq := sq.StatementBuilder.Select("*").From("gtfs_feed_infos").Where(sq.Eq{"feed_version_id": fvid}).OrderBy("feed_start_date").Limit(1)
	find.MustSelect(finder.DBX(), fiq, &fis)

	// Setup
	fetched := ents[0].FetchedAt.Time // time.Parse("2006-01-02", "2018-08-13")
	fmt.Println("fetched:", fetched)

	// Check if we have feed infos, otherwise calculate based on fetched week
	if len(fis) > 0 {
		fi := fis[0]
		fmt.Println("using feed info:", fi.FeedStartDate, fi.FeedEndDate)
		startDate = fi.FeedStartDate.Time
		endDate = fi.FeedEndDate.Time
	} else {
		// Get the week including fetched_at and the highest service week
		fetchedWeek := -1
		highestIdx := 0
		highestService := 0
		for i, ent := range ents {
			sd := ent.StartDate.Time
			ed := ent.EndDate.Time
			if (sd.Before(fetched) || sd.Equal(fetched)) && (ed.After(fetched) || ed.Equal(fetched)) {
				fetchedWeek = i
			}
			if ent.TotalService.Int > highestService {
				highestIdx = i
				highestService = ent.TotalService.Int
			}
		}
		if fetchedWeek < 0 {
			fmt.Println("using highest week:", highestIdx, highestService)
			fetchedWeek = highestIdx
		} else {
			fmt.Println("using fetched week:", fetchedWeek)
		}

		// Expand window in both directions from chosen week
		startDate = ents[fetchedWeek].StartDate.Time
		endDate = ents[fetchedWeek].EndDate.Time
		for i := fetchedWeek; i < len(ents); i++ {
			ent := ents[i]
			if float64(ent.TotalService.Int)/float64(highestService) < minServiceRatio {
				break
			}
			if ent.StartDate.Time.Before(startDate) {
				startDate = ent.StartDate.Time
			}
			endDate = ent.EndDate.Time
		}
		for i := fetchedWeek - 1; i > 0; i-- {
			ent := ents[i]
			if float64(ent.TotalService.Int)/float64(highestService) < minServiceRatio {
				break
			}
			if ent.EndDate.Time.After(endDate) {
				endDate = ent.EndDate.Time
			}
			startDate = ent.StartDate.Time
		}
	}

	// Check again to find the highest service week in the window
	fmt.Println("start:", startDate)
	fmt.Println("end:", endDate)
	bestWeek = startDate
	bestService := 0
	for _, ent := range ents {
		sd := ent.StartDate.Time
		ed := ent.EndDate.Time
		if (sd.Before(fetched) || sd.Equal(fetched)) && (ed.After(fetched) || ed.Equal(fetched)) {
			if ent.TotalService.Int > bestService {
				bestService = ent.TotalService.Int
				bestWeek = ent.StartDate.Time
			}
		}
	}
	fmt.Println("best service week:", bestWeek)
	return startDate, endDate, bestWeek, nil
}
