package resolvers

import (
	"encoding/json"
	"os"
	"testing"

	"github.com/99designs/gqlgen/client"
	"github.com/interline-io/transitland-lib/log"
	"github.com/interline-io/transitland-server/config"
	"github.com/interline-io/transitland-server/find"
	"github.com/interline-io/transitland-server/internal/clock"
	"github.com/interline-io/transitland-server/internal/gbfsfinder"
	"github.com/interline-io/transitland-server/internal/rtfinder"
	"github.com/interline-io/transitland-server/model"
	"github.com/stretchr/testify/assert"
	"github.com/tidwall/gjson"
)

var TestDBFinder model.Finder
var TestRTFinder model.RTFinder
var TestGbfsFinder model.GbfsFinder

func TestMain(m *testing.M) {
	g := os.Getenv("TL_TEST_SERVER_DATABASE_URL")
	if g == "" {
		log.Print("TL_TEST_SERVER_DATABASE_URL not set, skipping")
		return
	}
	db := find.MustOpenDB(g)
	dbf := find.NewDBFinder(db)
	TestDBFinder = dbf
	TestRTFinder = rtfinder.NewFinder(rtfinder.NewLocalCache(), db)
	TestGbfsFinder = gbfsfinder.NewFinder(nil)
	os.Exit(m.Run())
}

// Test helpers

func newTestClient() *client.Client {
	rtf := rtfinder.NewFinder(rtfinder.NewLocalCache(), TestDBFinder.DBX())
	cfg := config.Config{}
	srv, _ := NewServer(cfg, TestDBFinder, rtf, TestGbfsFinder)
	return client.New(srv)
}

func newTestClientWithClock(cl clock.Clock) (model.Finder, model.RTFinder, *client.Client) {
	// Create a new finder, with specified time
	cfg := config.Config{Clock: cl}
	db := TestDBFinder.DBX()
	rtf := rtfinder.NewFinder(rtfinder.NewLocalCache(), db)
	rtf.Clock = cl
	dbf := find.NewDBFinder(db)
	dbf.Clock = cl
	srv, _ := NewServer(cfg, dbf, rtf, nil)
	return dbf, rtf, client.New(srv)
}

func toJson(m map[string]interface{}) string {
	rr, _ := json.Marshal(&m)
	return string(rr)
}

type hw = map[string]interface{}

type testcase struct {
	name         string
	query        string
	vars         hw
	expect       string
	selector     string
	expectSelect []string
}

func testquery(t *testing.T, c *client.Client, tc testcase) {
	var resp map[string]interface{}
	opts := []client.Option{}
	for k, v := range tc.vars {
		opts = append(opts, client.Var(k, v))
	}
	c.MustPost(tc.query, &resp, opts...)
	jj := toJson(resp)
	if tc.expect != "" {
		if !assert.JSONEq(t, tc.expect, jj) {
			t.Errorf("got %s -- expect %s\n", jj, tc.expect)
		}
	}
	if tc.selector != "" {
		a := []string{}
		for _, v := range gjson.Get(jj, tc.selector).Array() {
			a = append(a, v.String())
		}
		if len(a) == 0 && tc.expectSelect == nil {
			t.Errorf("selector '%s' returned zero elements", tc.selector)
		} else {
			if !assert.ElementsMatch(t, tc.expectSelect, a) {
				t.Errorf("got %#v -- expect %#v\n\n", a, tc.expectSelect)
			}
		}
	}
}
