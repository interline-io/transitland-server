package authz

import (
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/interline-io/transitland-server/auth"
)

func filterTestTuple(tks []TestTupleKey, testType string, objectType ObjectType, hasAction Action) []TestTupleKey {
	var ret []TestTupleKey
	for _, tk := range tks {
		if tk.Test != testType {
			continue
		}
		if tk.Object.Type != objectType {
			continue
		}
		match := false
		if hasAction == 0 {
			match = true
		}
		for _, checkAction := range tk.Checks {
			if checkAction == hasAction.String() {
				match = true
			}
		}
		if !match {
			continue
		}
		ret = append(ret, tk)
	}
	return ret
}

func TestServer(t *testing.T) {
	// Test assertions
	checks, err := LoadTuples("../test/authz/tls.csv")
	if err != nil {
		t.Fatal(err)
	}

	// TENANTS
	t.Run("TenantList", func(t *testing.T) {
		checker := newTestChecker(t)
		for _, tk := range filterTestTuple(checks, "list", TenantType, CanView) {
			t.Run(tk.String(), func(t *testing.T) {
				srv := testServerWithUser(checker, tk)
				req, _ := http.NewRequest("GET", "/tenants", nil)
				rr := httptest.NewRecorder()
				srv.ServeHTTP(rr, req)
				checkHttpExpectError(t, tk, rr)
			})
		}
	})

	t.Run("TenantPermissions", func(t *testing.T) {
		checker := newTestChecker(t)
		for _, tk := range filterTestTuple(checks, "check", TenantType, CanView) {
			t.Run(tk.String(), func(t *testing.T) {
				srv := testServerWithUser(checker, tk)
				req, _ := http.NewRequest("GET", fmt.Sprintf("/tenants/%s", tk.Object.Name), nil)
				rr := httptest.NewRecorder()
				srv.ServeHTTP(rr, req)
				checkHttpExpectError(t, tk, rr)
			})
		}
	})

	// GROUPS
	t.Run("GroupList", func(t *testing.T) {
		checker := newTestChecker(t)
		for _, tk := range filterTestTuple(checks, "list", GroupType, CanView) {
			t.Run(tk.String(), func(t *testing.T) {
				srv := testServerWithUser(checker, tk)
				req, _ := http.NewRequest("GET", "/groups", nil)
				rr := httptest.NewRecorder()
				srv.ServeHTTP(rr, req)
				checkHttpExpectError(t, tk, rr)
			})
		}
	})

	t.Run("GroupPermissions", func(t *testing.T) {
		checker := newTestChecker(t)
		for _, tk := range filterTestTuple(checks, "check", GroupType, CanView) {
			t.Run(tk.String(), func(t *testing.T) {
				srv := testServerWithUser(checker, tk)
				req, _ := http.NewRequest("GET", fmt.Sprintf("/groups/%s", tk.Object.Name), nil)
				rr := httptest.NewRecorder()
				srv.ServeHTTP(rr, req)
				checkHttpExpectError(t, tk, rr)
			})
		}
	})

	// FEEDS
	t.Run("FeedList", func(t *testing.T) {
		checker := newTestChecker(t)
		for _, tk := range filterTestTuple(checks, "list", FeedType, CanView) {
			t.Run(tk.String(), func(t *testing.T) {
				srv := testServerWithUser(checker, tk)
				req, _ := http.NewRequest("GET", "/feeds", nil)
				rr := httptest.NewRecorder()
				srv.ServeHTTP(rr, req)
				checkHttpExpectError(t, tk, rr)
			})
		}
	})

	t.Run("FeedPermissions", func(t *testing.T) {
		checker := newTestChecker(t)
		for _, tk := range filterTestTuple(checks, "check", FeedType, CanView) {
			t.Run(tk.String(), func(t *testing.T) {
				srv := testServerWithUser(checker, tk)
				req, _ := http.NewRequest("GET", fmt.Sprintf("/feeds/%s", tk.Object.Name), nil)
				rr := httptest.NewRecorder()
				srv.ServeHTTP(rr, req)
				checkHttpExpectError(t, tk, rr)
			})
		}
	})

	// FEED VERSIONS
	t.Run("FeedVersionList", func(t *testing.T) {
		checker := newTestChecker(t)
		for _, tk := range filterTestTuple(checks, "list", FeedVersionType, CanView) {
			t.Run(tk.String(), func(t *testing.T) {
				srv := testServerWithUser(checker, tk)
				req, _ := http.NewRequest("GET", "/feed_versions", nil)
				rr := httptest.NewRecorder()
				srv.ServeHTTP(rr, req)
				checkHttpExpectError(t, tk, rr)
			})
		}
	})

	t.Run("FeedVersionPermissions", func(t *testing.T) {
		checker := newTestChecker(t)
		for _, tk := range filterTestTuple(checks, "check", FeedVersionType, CanView) {
			t.Run(tk.String(), func(t *testing.T) {
				srv := testServerWithUser(checker, tk)
				req, _ := http.NewRequest("GET", fmt.Sprintf("/feed_versions/%s", tk.Object.Name), nil)
				rr := httptest.NewRecorder()
				srv.ServeHTTP(rr, req)
				checkHttpExpectError(t, tk, rr)
			})
		}
	})

}

func testServerWithUser(c *Checker, tk TestTupleKey) http.Handler {
	srv, _ := NewServer(c)
	srv = auth.UserDefaultMiddleware(stringOr(tk.CheckAsUser, tk.Subject.Name))(srv)
	return srv
}

func printHttpResponse(t testing.TB, r io.Reader) {
	b, _ := ioutil.ReadAll(r)
	t.Log(string(b))
}

func checkHttpExpectError(t testing.TB, tk TestTupleKey, rr *httptest.ResponseRecorder) {
	status := rr.Code
	if tk.ExpectErrorAsUser && status == http.StatusOK {
		t.Errorf("got error code %d, expected non-200", status)
		printHttpResponse(t, rr.Body)
	} else if !tk.ExpectErrorAsUser && status != http.StatusOK {
		t.Errorf("got error code %d, expected 200", status)
		printHttpResponse(t, rr.Body)
	} else {
		printHttpResponse(t, rr.Body)
	}
}
