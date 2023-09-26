package gql

import (
	"encoding/json"
	"log"
	"os"
	"testing"
	"time"

	"github.com/99designs/gqlgen/client"
	"github.com/interline-io/transitland-server/auth/ancheck"
	"github.com/interline-io/transitland-server/internal/clock"
	"github.com/interline-io/transitland-server/internal/testfinder"
	"github.com/interline-io/transitland-server/internal/testutil"
	"github.com/interline-io/transitland-server/model"
	"github.com/stretchr/testify/assert"
	"github.com/tidwall/gjson"
)

type hw = map[string]interface{}

type testcase struct {
	name               string
	query              string
	vars               hw
	expect             string
	user               string
	selector           string
	expectError        bool
	selectExpect       []string
	selectExpectUnique []string
	selectExpectCount  int
	f                  func(*testing.T, string)
}

func TestMain(m *testing.M) {
	// Increase default limit for testing purposes
	MAXLIMIT = 100_000
	if a, ok := testutil.CheckTestDB(); !ok {
		log.Print(a)
		return
	}
	os.Exit(m.Run())
}

// Test helpers

func newTestClient(t testing.TB) (*client.Client, model.Finders) {
	when, err := time.Parse("2006-01-02T15:04:05", "2022-09-01T00:00:00")
	if err != nil {
		t.Fatal(err)
	}
	return newTestClientWithClock(t, &clock.Mock{T: when}, testfinder.DefaultRTJson())
}

func newTestClientWithClock(t testing.TB, cl clock.Clock, rtfiles []testfinder.RTJsonFile) (*client.Client, model.Finders) {
	te := testfinder.Finders(t, cl, rtfiles)
	srv, _ := NewServer(te.Config, te.Finder, te.RTFinder, te.GbfsFinder, te.Checker)
	srvMiddleware := ancheck.UserDefaultMiddleware("testuser")
	return client.New(srvMiddleware(srv)), te
}

func toJson(m map[string]interface{}) string {
	rr, _ := json.Marshal(&m)
	return string(rr)
}

func queryTestcase(t *testing.T, c *client.Client, tc testcase) {
	tested := false
	var resp map[string]interface{}
	opts := []client.Option{}
	for k, v := range tc.vars {
		opts = append(opts, client.Var(k, v))
	}
	if err := c.Post(tc.query, &resp, opts...); err != nil {
		if tc.expectError {
			// ok
		} else {
			t.Error(err)
			return
		}
	} else if tc.expectError {
		t.Error("expected error")
	}
	jj := toJson(resp)
	if tc.expect != "" {
		tested = true
		if !assert.JSONEq(t, tc.expect, jj) {
			t.Errorf("got %s -- expect %s\n", jj, tc.expect)
		}
	}
	if tc.f != nil {
		tested = true
		tc.f(t, jj)
	}
	if tc.selector != "" {
		a := []string{}
		for _, v := range gjson.Get(jj, tc.selector).Array() {
			a = append(a, v.String())
		}
		if tc.selectExpectCount != 0 {
			tested = true
			if len(a) != tc.selectExpectCount {
				t.Errorf("selector returned %d elements, expected %d", len(a), tc.selectExpectCount)
			}
		}
		if tc.selectExpectUnique != nil {
			tested = true
			mm := map[string]int{}
			for _, v := range a {
				mm[v] += 1
			}
			var keys []string
			for k := range mm {
				keys = append(keys, k)
			}
			assert.ElementsMatch(t, tc.selectExpectUnique, keys)
		}
		if tc.selectExpect != nil {
			tested = true
			if !assert.ElementsMatch(t, tc.selectExpect, a) {
				t.Errorf("got %#v -- expect %#v\n\n", a, tc.selectExpect)
			}
		}
	}
	if !tested {
		t.Errorf("no test performed, check test case")
	}
}

func queryTestcases(t *testing.T, c *client.Client, tcs []testcase) {
	for _, tc := range tcs {
		t.Run(tc.name, func(t *testing.T) {
			queryTestcase(t, c, tc)
		})
	}
}

func benchmarkTestcases(b *testing.B, c *client.Client, tcs []testcase) {
	for _, tc := range tcs {
		b.Run(tc.name, func(b *testing.B) {
			benchmarkTestcase(b, c, tc)
		})
	}
}

func benchmarkTestcase(b *testing.B, c *client.Client, tc testcase) {
	opts := []client.Option{}
	for k, v := range tc.vars {
		opts = append(opts, client.Var(k, v))
	}
	var resp map[string]any
	b.ResetTimer()
	for n := 0; n < b.N; n++ {
		c.MustPost(tc.query, &resp, opts...)
	}
}

func Test_checkBbox(t *testing.T) {
	tcs := []struct {
		name   string
		bbox   model.BoundingBox
		area   float64
		expect bool
	}{
		{
			name:   "small 1",
			bbox:   model.BoundingBox{MinLon: -122.27023950636408, MinLat: 37.80659174954722, MaxLon: -122.26747210439623, MaxLat: 37.80928736232528},
			area:   0.08 * 1e6,
			expect: true,
		},
		{
			name:   "small 2",
			bbox:   model.BoundingBox{MinLon: -122.27023950636408, MinLat: 37.80659174954722, MaxLon: -122.26747210439623, MaxLat: 37.80928736232528},
			area:   0.07 * 1e6,
			expect: false,
		},
		{
			name:   "big 1",
			bbox:   model.BoundingBox{MinLon: -123.08182748515924, MinLat: 37.100203650623826, MaxLon: -121.31739022765265, MaxLat: 38.31972646345332},
			area:   22000.0 * 1e6,
			expect: true,
		},
		{
			name:   "big 2",
			bbox:   model.BoundingBox{MinLon: -123.08182748515924, MinLat: 37.100203650623826, MaxLon: -121.31739022765265, MaxLat: 38.31972646345332},
			area:   20000.0 * 1e6,
			expect: false,
		},
	}
	for _, tc := range tcs {
		t.Run(tc.name, func(t *testing.T) {
			v := checkBbox(&tc.bbox, tc.area)
			assert.Equal(t, tc.expect, v, "expected result")

		})
	}
}
