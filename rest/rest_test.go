package rest

import (
	"context"
	"encoding/json"
	"os"
	"testing"

	"github.com/interline-io/transitland-lib/log"
	"github.com/interline-io/transitland-server/config"
	"github.com/interline-io/transitland-server/find"
	"github.com/interline-io/transitland-server/internal/rtcache"
	"github.com/interline-io/transitland-server/model"
	"github.com/interline-io/transitland-server/resolvers"
	"github.com/stretchr/testify/assert"
	"github.com/tidwall/gjson"
)

var TestDBFinder model.Finder
var TestRTFinder model.RTFinder

const LON = 37.803613
const LAT = -122.271556

func TestMain(m *testing.M) {
	g := os.Getenv("TL_TEST_SERVER_DATABASE_URL")
	if g == "" {
		log.Print("TL_TEST_SERVER_DATABASE_URL not set, skipping")
		return
	}
	db := find.MustOpenDB(g)
	TestDBFinder = find.NewDBFinder(db)
	TestRTFinder = rtcache.NewRTFinder(rtcache.NewLocalCache(), db)
	os.Exit(m.Run())
}

// Test helpers

func testRestConfig() restConfig {
	cfg := config.Config{}
	srv, _ := resolvers.NewServer(cfg, TestDBFinder, TestRTFinder)
	return restConfig{srv: srv, Config: cfg}
}

func toJson(m map[string]interface{}) string {
	rr, _ := json.Marshal(&m)
	return string(rr)
}

type testRest struct {
	name         string
	h            apiHandler
	format       string
	selector     string
	expectSelect []string
	expectLength int
}

func testquery(t *testing.T, cfg restConfig, tc testRest) {
	data, err := makeRequest(context.TODO(), cfg, tc.h, tc.format, nil)
	if err != nil {
		t.Error(err)
		return
	}
	jj := string(data)
	if tc.selector != "" {
		a := []string{}
		for _, v := range gjson.Get(jj, tc.selector).Array() {
			a = append(a, v.String())
		}
		if len(tc.expectSelect) > 0 {
			if len(a) == 0 {
				t.Errorf("selector '%s' returned zero elements", tc.selector)
			} else {
				if !assert.ElementsMatch(t, a, tc.expectSelect) {
					t.Errorf("got %#v -- expect %#v\n\n", a, tc.expectSelect)
				}
			}
		} else {
			if len(a) != tc.expectLength {
				t.Errorf("got %d elements, expected %d", len(a), tc.expectLength)
			}
		}
	}
}
