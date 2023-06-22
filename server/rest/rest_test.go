package rest

import (
	"context"
	"encoding/json"
	"net/http"
	"os"
	"testing"
	"time"

	"github.com/interline-io/transitland-lib/log"
	"github.com/interline-io/transitland-server/config"
	"github.com/interline-io/transitland-server/finders/dbfinder"
	"github.com/interline-io/transitland-server/internal/clock"
	"github.com/interline-io/transitland-server/internal/testfinder"
	"github.com/interline-io/transitland-server/server/gql"
	"github.com/stretchr/testify/assert"
	"github.com/tidwall/gjson"
)

func TestMain(m *testing.M) {
	dbfinder.MAXLIMIT = 100_000
	MAXLIMIT = dbfinder.MAXLIMIT
	g := os.Getenv("TL_TEST_SERVER_DATABASE_URL")
	if g == "" {
		log.Print("TL_TEST_SERVER_DATABASE_URL not set, skipping")
		return
	}
	os.Exit(m.Run())
}

func testRestConfig(t testing.TB) (http.Handler, testfinder.TestEnv) {
	when, err := time.Parse("2006-01-02T15:04:05", "2018-06-01T00:00:00")
	if err != nil {
		t.Fatal(err)
	}
	te := testfinder.Finders(t, &clock.Mock{T: when}, testfinder.DefaultRTJson())
	srv, err := gql.NewServer(te.Config, te.Finder, te.RTFinder, te.GbfsFinder, nil)
	if err != nil {
		panic(err)
	}
	return srv, te
}

func testRestServer(t testing.TB, cfg config.Config, srv http.Handler) (http.Handler, error) {
	return NewServer(cfg, srv)
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
	f            func(*testing.T, string)
}

func testquery(t *testing.T, srv http.Handler, te testfinder.TestEnv, tc testRest) {
	data, err := makeRequest(context.TODO(), restConfig{srv: srv, Config: te.Config}, tc.h, tc.format, nil)
	if err != nil {
		t.Error(err)
		return
	}
	jj := string(data)
	tested := false
	if tc.f != nil {
		tested = true
		tc.f(t, jj)
	}
	if tc.selector != "" {
		tested = true
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
	if !tested {
		t.Errorf("no test performed, check test case")
	}
}
