package rest

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"os"
	"testing"

	"github.com/interline-io/transitland-dbutil/testutil"
	"github.com/interline-io/transitland-server/internal/testconfig"
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

type testCase struct {
	name         string
	h            apiHandler
	format       string
	selector     string
	expectSelect []string
	expectLength int
	f            func(*testing.T, string)
}

func testHandlersWithOptions(t testing.TB, opts testconfig.Options) (http.Handler, http.Handler, model.Config) {
	cfg := testconfig.Config(t, opts)
	graphqlHandler, err := gql.NewServer()
	if err != nil {
		t.Fatal(err)
	}
	restHandler, err := NewServer(graphqlHandler)
	if err != nil {
		t.Fatal(err)
	}
	return model.AddConfigAndPerms(cfg, graphqlHandler),
		model.AddConfigAndPerms(cfg, restHandler),
		cfg
}

func checkTestCase(t *testing.T, tc testCase) {
	opts := testconfig.Options{
		When:    "2018-06-01T00:00:00",
		RTJsons: testconfig.DefaultRTJson(),
	}
	cfg := testconfig.Config(t, opts)
	graphqlHandler, err := gql.NewServer()
	if err != nil {
		t.Fatal(err)
	}
	restHandler, err := NewServer(graphqlHandler)
	if err != nil {
		t.Fatal(err)
	}
	checkTestCaseWithHandlers(
		t,
		tc,
		model.AddConfigAndPerms(cfg, graphqlHandler),
		model.AddConfigAndPerms(cfg, restHandler),
	)
}

func checkTestCaseWithHandlers(t *testing.T, tc testCase, graphqlHandler http.Handler, restHandler http.Handler) {
	data, err := makeRequest(context.TODO(), graphqlHandler, tc.h, tc.format, nil)
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

func toJson(m map[string]interface{}) string {
	rr, _ := json.Marshal(&m)
	return string(rr)
}
