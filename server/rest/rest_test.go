package rest

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"os"
	"testing"

	"github.com/interline-io/transitland-server/internal/testconfig"
	"github.com/interline-io/transitland-server/internal/testutil"
	"github.com/interline-io/transitland-server/model"
	"github.com/interline-io/transitland-server/server/gql"
	"github.com/stretchr/testify/assert"
	"github.com/tidwall/gjson"
)

func TestMain(m *testing.M) {
	// Increase limit for test
	MAXLIMIT = 100_000
	gql.MAXLIMIT = MAXLIMIT
	if a, ok := testutil.CheckTestDB(); !ok {
		log.Print(a)
		return
	}
	os.Exit(m.Run())
}

func testRestConfig(t testing.TB) (http.Handler, model.Config) {
	cfg := testconfig.Config(t,
		testconfig.Options{
			When:    "2018-06-01T00:00:00",
			RTJsons: testconfig.DefaultRTJson(),
		},
	)
	srv, err := gql.NewServer(cfg)
	if err != nil {
		panic(err)
	}
	srv = model.AddConfig(cfg)(srv)
	return srv, cfg
}

func testRestServer(t testing.TB, cfg Config, srv http.Handler) (http.Handler, error) {
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

func testquery(t *testing.T, graphqlHandler http.Handler, tc testRest) {
	data, err := makeRequest(context.TODO(), Config{}, graphqlHandler, tc.h, tc.format, nil)
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
