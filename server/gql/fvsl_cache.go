package gql

import (
	"context"
	"sync"
	"time"

	"github.com/interline-io/log"
	"github.com/interline-io/transitland-server/model"
)

type fvslWindow struct {
	FetchedAt time.Time
	StartDate time.Time
	EndDate   time.Time
	BestWeek  time.Time
	Valid     bool
}

type fvslCache struct {
	Finder    model.Finder
	lock      sync.Mutex
	fvWindows map[int]fvslWindow
}

func newFvslCache(f model.Finder) *fvslCache {
	return &fvslCache{
		Finder:    f,
		fvWindows: map[int]fvslWindow{},
	}
}

func (f *fvslCache) Get(fvid int) (fvslWindow, bool) {
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

func (f *fvslCache) Set(fvid int, w fvslWindow) {
	f.lock.Lock()
	defer f.lock.Unlock()
	f.fvWindows[fvid] = w
}

func (f *fvslCache) query(fvid int) (fvslWindow, error) {
	var err error
	w := fvslWindow{}
	w.StartDate, w.EndDate, w.BestWeek, err = f.Finder.FindFeedVersionServiceWindow(context.TODO(), fvid)
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
