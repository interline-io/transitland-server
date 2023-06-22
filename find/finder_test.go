package find

import (
	"context"
	"log"
	"os"
	"testing"

	"github.com/interline-io/transitland-server/model"
	"github.com/stretchr/testify/assert"
)

var TestDBFinder model.Finder

func TestMain(m *testing.M) {
	g := os.Getenv("TL_TEST_SERVER_DATABASE_URL")
	if g == "" {
		log.Print("TL_TEST_SERVER_DATABASE_URL not set, skipping")
		return
	}
	db := MustOpenDB(g)
	dbf := NewDBFinder(db)
	TestDBFinder = dbf
	os.Exit(m.Run())
}

func TestFinder_FindFeedVersionServiceWindow(t *testing.T) {
	fvm := map[string]int{}
	fvs, err := TestDBFinder.FindFeedVersions(context.TODO(), nil, nil, nil, nil, nil)
	if err != nil {
		t.Fatal(err)
	}
	for _, fv := range fvs {
		fvm[fv.SHA1] = fv.ID
	}
	tcs := []struct {
		name  string
		fvid  int
		start string
		end   string
		best  string
	}{
		{
			"hart",
			fvm["c969427f56d3a645195dd8365cde6d7feae7e99b"],
			"2018-02-26", // calculated
			"2018-10-21",
			"2018-07-09",
		},
		{
			"bart",
			fvm["e535eb2b3b9ac3ef15d82c56575e914575e732e0"],
			"2018-05-26", // from feed info
			"2019-07-01", // from feed info
			"2018-06-04",
		},
		{
			"caltrain",
			fvm["d2813c293bcfd7a97dde599527ae6c62c98e66c6"],
			"2018-06-18", // calculated
			"2019-10-06",
			"2018-06-18",
		},
	}
	for _, tc := range tcs {
		t.Run(tc.name, func(t *testing.T) {
			start, end, best, err := TestDBFinder.FindFeedVersionServiceWindow(context.TODO(), tc.fvid)
			if err != nil {
				t.Fatal(err)
			}
			df := "2006-01-02"
			assert.Equal(t, tc.start, start.Format(df), "did not get expected window start")
			assert.Equal(t, tc.end, end.Format(df), "did not get expected window end")
			assert.Equal(t, tc.best, best.Format(df), "did not get expected best week in window")
			if end.Before(end) {
				t.Errorf("window end date %s before window start date %s", start.Format(df), end.Format(df))
			}
			if best.Before(start) {
				t.Errorf("best date %s before window start date %s", best.Format(df), start.Format(df))
			}
			if best.After(end) {
				t.Errorf("best date %s after window end date %s", best.Format(df), end.Format(df))
			}
		})
	}
}
