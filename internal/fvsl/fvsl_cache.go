package fvsl

import (
	"sync"
	"time"

	"github.com/interline-io/transitland-server/model"
	"github.com/rs/zerolog/log"
)

type FVSLWindow struct {
	FetchedAt time.Time
	StartDate time.Time
	EndDate   time.Time
	BestWeek  time.Time
	Valid     bool
}

type FVSLCache struct {
	Finder    model.Finder
	lock      sync.Mutex
	fvWindows map[int]FVSLWindow
}

func (f *FVSLCache) Get(fvid int) (FVSLWindow, bool) {
	f.lock.Lock()
	a, ok := f.fvWindows[fvid]
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
	if f.fvWindows == nil {
		f.fvWindows = map[int]FVSLWindow{}
	}
	f.fvWindows[fvid] = w
}

func (f *FVSLCache) query(fvid int) (FVSLWindow, error) {
	var err error
	w := FVSLWindow{}
	w.StartDate, w.EndDate, w.BestWeek, err = f.Finder.FindFeedVersionServiceWindow(fvid)
	log.Trace().
		Str("start_date", w.StartDate.Format("2006-01-02")).
		Str("end_date", w.EndDate.Format("2006-01-02")).
		Str("best_week", w.BestWeek.Format("2006-01-02")).
		Int("fvid", fvid).
		Msg("service window result")
	if err != nil {
		return w, err
	}
	w.Valid = true
	return w, err
}
