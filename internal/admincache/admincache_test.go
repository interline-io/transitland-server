package admincache

import (
	"context"
	"testing"

	sq "github.com/Masterminds/squirrel"
	"github.com/interline-io/transitland-server/internal/dbutil"
	"github.com/interline-io/transitland-server/internal/testutil"
	"github.com/interline-io/transitland-server/internal/xy"
	"github.com/stretchr/testify/assert"
)

type testCase struct {
	name        string
	point       xy.Point
	expectAdm0  string
	expectAdm1  string
	expectCount int
	skipPg      bool
}

// Note: these values are based on the Natural Earth 10m data set, which is slightly simplified. For instance, the Georgia/Florida boundary used below.
func getTestCases() []testCase {
	tcs := []testCase{
		{name: "new york", expectAdm0: "United States of America", expectAdm1: "New York", expectCount: 1, point: xy.Point{Lon: -74.132285, Lat: 40.625665}},
		{name: "california", expectAdm0: "United States of America", expectAdm1: "California", expectCount: 1, point: xy.Point{Lon: -122.431297, Lat: 37.773972}},
		{name: "kansas 1", expectAdm0: "United States of America", expectAdm1: "Kansas", expectCount: 1, point: xy.Point{Lon: -98.85867269364557, Lat: 39.96773433000109}},
		{name: "kansas 2", expectAdm0: "United States of America", expectAdm1: "Kansas", expectCount: 1, point: xy.Point{Lon: -98.85867269364557, Lat: 39.99901}},
		{name: "nebraska 1", expectAdm0: "United States of America", expectAdm1: "Nebraska", expectCount: 1, point: xy.Point{Lon: -98.862255, Lat: 40.001587}},
		{name: "nebraska 2", expectAdm0: "United States of America", expectAdm1: "Nebraska", expectCount: 1, point: xy.Point{Lon: -98.867745, Lat: 40.003185}},
		{name: "utah", expectAdm0: "United States of America", expectAdm1: "Utah", expectCount: 1, point: xy.Point{Lon: -109.056664, Lat: 40.996479}},
		{name: "colorado", expectAdm0: "United States of America", expectAdm1: "Colorado", expectCount: 1, point: xy.Point{Lon: -109.045685, Lat: 40.997833}},
		{name: "wyoming", expectAdm0: "United States of America", expectAdm1: "Wyoming", expectCount: 1, point: xy.Point{Lon: -109.050133, Lat: 41.002209}},
		{name: "north dakota", expectAdm0: "United States of America", expectAdm1: "North Dakota", expectCount: 1, point: xy.Point{Lon: -100.964531, Lat: 45.946934}},
		{name: "georgia", expectAdm0: "United States of America", expectAdm1: "Georgia", expectCount: 1, point: xy.Point{Lon: -82.066697, Lat: 30.370054}},
		{name: "florida", expectAdm0: "United States of America", expectAdm1: "Florida", expectCount: 1, point: xy.Point{Lon: -82.046522, Lat: 30.360419}},
		{name: "saskatchewan", expectAdm0: "Canada", expectAdm1: "Saskatchewan", expectCount: 1, point: xy.Point{Lon: -102.007904, Lat: 58.269615}},
		{name: "manitoba", expectAdm0: "Canada", expectAdm1: "Manitoba", expectCount: 1, point: xy.Point{Lon: -101.982025, Lat: 58.269245}},
		{name: "paris", expectAdm0: "France", expectAdm1: "Paris", expectCount: 1, point: xy.Point{Lon: 2.4729377, Lat: 48.8589143}},
		{name: "texas", expectAdm0: "United States of America", expectAdm1: "Texas", expectCount: 1, point: xy.Point{Lon: -94.794261, Lat: 29.289210}},
		{name: "texas water 1", skipPg: true, expectAdm0: "United States of America", expectAdm1: "Texas", expectCount: 1, point: xy.Point{Lon: -94.784667, Lat: 29.286234}},
		// {name: "texas water 1", point: xy.Point{Lon: -94.784667, Lat: 29.286234}},
		{name: "texas water 2", expectCount: 0, point: xy.Point{Lon: -94.237, Lat: 26.874}},
		{name: "null", expectCount: 0, point: xy.Point{Lon: 0, Lat: 0}},
	}
	return tcs
}

func TestAdminCache(t *testing.T) {
	dbx := testutil.MustOpenTestDB()
	c := NewAdminCache()
	c.LoadAdmins(context.Background(), dbx)
	tcs := getTestCases()
	for _, tc := range tcs {
		t.Run(tc.name, func(t *testing.T) {
			r := c.Check(tc.point)
			assert.Equal(t, tc.expectAdm0, r.Adm0Name)
			assert.Equal(t, tc.expectAdm1, r.Adm1Name)
			assert.Equal(t, tc.expectCount, r.Count)
			if tc.skipPg {
				return
			}
			var pgCheck []struct {
				Name  string
				Admin string
			}
			q := sq.
				Select("ne.name", "ne.admin", "ne.geometry").
				From("ne_10m_admin_1_states_provinces ne").
				Where("ST_Intersects(ne.geometry::geography, ST_MakePoint(?,?)::geography)", tc.point.Lon, tc.point.Lat)
			if err := dbutil.Select(context.Background(), dbx, q, &pgCheck); err != nil {
				t.Fatal(err)
			}
			if len(pgCheck) != tc.expectCount {
				t.Error("expectCount did not match result from postgres")
			}
			for _, ent := range pgCheck {
				assert.Equal(t, tc.expectAdm0, ent.Admin, "different than postgres")
				assert.Equal(t, tc.expectAdm1, ent.Name, "different than postgres")
			}
		})
	}
}

func BenchmarkTestAdminCache(b *testing.B) {
	dbx := testutil.MustOpenTestDB()
	c := NewAdminCache()
	c.LoadAdmins(context.Background(), dbx)
	b.ResetTimer()
	tcs := getTestCases()
	for _, tc := range tcs {
		b.Run(tc.name, func(b *testing.B) {
			for n := 0; n < b.N; n++ {
				r := c.Check(tc.point)
				_ = r
			}
		})
	}
}

func BenchmarkTestAdminCache_LoadAdmins(b *testing.B) {
	dbx := testutil.MustOpenTestDB()
	c := NewAdminCache()
	for n := 0; n < b.N; n++ {
		if err := c.LoadAdmins(context.Background(), dbx); err != nil {
			b.Fatal(err)
		}
	}
}
