package rest

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"testing"

	"github.com/interline-io/transitland-lib/log"
	"github.com/interline-io/transitland-lib/rt/pb"
	"github.com/interline-io/transitland-server/config"
	"github.com/interline-io/transitland-server/find"
	"github.com/interline-io/transitland-server/internal/rtfinder"
	"github.com/interline-io/transitland-server/internal/testutil"
	"github.com/interline-io/transitland-server/model"
	"github.com/interline-io/transitland-server/resolvers"
	"github.com/stretchr/testify/assert"
	"github.com/tidwall/gjson"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"
)

var TestDBFinder model.Finder
var TestRTFinder model.RTFinder

const LON = 37.803613
const LAT = -122.271556

func TestMain(m *testing.M) {
	find.MAXLIMIT = 100_000
	MAXLIMIT = 100_000
	g := os.Getenv("TL_TEST_SERVER_DATABASE_URL")
	if g == "" {
		log.Print("TL_TEST_SERVER_DATABASE_URL not set, skipping")
		return
	}
	db := find.MustOpenDB(g)
	TestDBFinder = find.NewDBFinder(db)
	TestRTFinder = rtfinder.NewFinder(rtfinder.NewLocalCache(), db)
	os.Exit(m.Run())
}

// Test helpers
type rtFile struct {
	feed  string
	ftype string
	fname string
}

func testRestConfig() restConfig {
	cfg := config.Config{}

	rtf := rtfinder.NewFinder(rtfinder.NewLocalCache(), TestDBFinder.DBX())
	var baseRTFiles = []rtFile{
		{"BA", "realtime_trip_updates", "BA.json"},
		{"CT", "realtime_trip_updates", "CT.json"},
	}
	for _, rtfile := range baseRTFiles {
		if err := rtFetchJson(rtfile.feed, rtfile.ftype, testutil.RelPath("test", "data", "rt", rtfile.fname), rtf); err != nil {
			panic(err)
		}
	}
	srv, _ := resolvers.NewServer(cfg, TestDBFinder, rtf, nil)
	return restConfig{srv: srv, Config: cfg}
}

func rtFetchJson(feed string, ftype string, url string, rtfinder model.RTFinder) error {
	var msg pb.FeedMessage
	jdata, err := ioutil.ReadFile(url)
	if err != nil {
		return err
	}
	if err := protojson.Unmarshal(jdata, &msg); err != nil {
		return err
	}
	rtdata, err := proto.Marshal(&msg)
	if err != nil {
		return err
	}
	key := fmt.Sprintf("rtdata:%s:%s", feed, ftype)
	return rtfinder.AddData(key, rtdata)
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
