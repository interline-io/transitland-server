package authz

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/interline-io/transitland-server/auth/authn"
	"github.com/interline-io/transitland-server/internal/dbutil"
	"github.com/interline-io/transitland-server/internal/generated/azpb"
	"github.com/stretchr/testify/assert"
	"github.com/tidwall/gjson"
)

func TestServer(t *testing.T) {
	fgaUrl, a, ok := dbutil.CheckEnv("TL_TEST_FGA_ENDPOINT")
	if !ok {
		t.Skip(a)
		return
	}
	dbx := dbutil.MustOpenTestDB()
	serverTestData := []TestTuple{
		{
			Subject:  NewEntityKey(TenantType, "tl-tenant"),
			Object:   NewEntityKey(GroupType, "BA-group"),
			Relation: ParentRelation,
			Notes:    "org:BA-group belongs to tenant:tl-tenant",
		},
		{
			Subject:  NewEntityKey(TenantType, "tl-tenant"),
			Object:   NewEntityKey(GroupType, "CT-group"),
			Relation: ParentRelation,
			Notes:    "org:CT-group belongs to tenant:tl-tenant",
		},
		{
			Subject:  NewEntityKey(TenantType, "restricted-tenant"),
			Object:   NewEntityKey(GroupType, "test-group"),
			Relation: ParentRelation,
		},
		{
			Subject:  NewEntityKey(UserType, "tl-tenant-admin"),
			Object:   NewEntityKey(TenantType, "tl-tenant"),
			Relation: AdminRelation,
		},

		{
			Subject:  NewEntityKey(GroupType, "BA-group"),
			Object:   NewEntityKey(FeedType, "BA"),
			Relation: ParentRelation,
		},
		{
			Subject:  NewEntityKey(GroupType, "CT-group"),
			Object:   NewEntityKey(FeedType, "CT"),
			Relation: ParentRelation,
		},

		{
			Subject:  NewEntityKey(UserType, "ian"),
			Object:   NewEntityKey(TenantType, "tl-tenant"),
			Relation: MemberRelation,
		},
		{
			Subject:  NewEntityKey(UserType, "drew"),
			Object:   NewEntityKey(TenantType, "tl-tenant"),
			Relation: MemberRelation,
		},
		{
			Subject:  NewEntityKey(UserType, "ian"),
			Object:   NewEntityKey(GroupType, "BA-group"),
			Relation: ViewerRelation,
		},
		{
			Subject:  NewEntityKey(UserType, "drew"),
			Object:   NewEntityKey(GroupType, "CT-group"),
			Relation: EditorRelation,
		},
		{
			Subject:  NewEntityKey(UserType, "drew"),
			Object:   NewEntityKey(FeedVersionType, "e535eb2b3b9ac3ef15d82c56575e914575e732e0"),
			Relation: ViewerRelation,
			Notes:    "assign drew permission to view this BA feed",
		},
	}

	// TENANTS
	t.Run("TenantList", func(t *testing.T) {
		checker := newTestChecker(t, fgaUrl, serverTestData)
		checks := []TestTuple{
			{
				Subject:    NewEntityKey(UserType, "tl-tenant-admin"),
				ExpectKeys: newEntityKeys(TenantType, "tl-tenant"),
			},
			{
				Subject:    NewEntityKey(UserType, "ian"),
				ExpectKeys: newEntityKeys(TenantType, "tl-tenant"),
			},
			{
				Subject:    NewEntityKey(UserType, "drew"),
				ExpectKeys: newEntityKeys(TenantType, "tl-tenant"),
			},
			{
				Subject:    NewEntityKey(UserType, "unknown"),
				ExpectKeys: newEntityKeys(TenantType),
			},
		}
		for _, tc := range checks {
			t.Run(tc.String(), func(t *testing.T) {
				srv := testServerWithUser(checker, tc)
				req, _ := http.NewRequest("GET", "/tenants", nil)
				rr := httptest.NewRecorder()
				srv.ServeHTTP(rr, req)
				checkHttpExpectError(t, tc, rr)
				assert.ElementsMatch(
					t,
					ekGetNames(tc.ExpectKeys),
					responseGetNames(t, rr.Body.Bytes(), "tenants", "name"),
				)
			})
		}
	})

	t.Run("TenantPermissions", func(t *testing.T) {
		checker := newTestChecker(t, fgaUrl, serverTestData)
		checks := []TestTuple{
			{
				Subject:       NewEntityKey(UserType, "tl-tenant-admin"),
				Object:        NewEntityKey(TenantType, "tl-tenant"),
				ExpectActions: []Action{CanView, CanEdit, CanEditMembers, CanCreateOrg, CanDeleteOrg},
			},
			{
				Subject:            NewEntityKey(UserType, "tl-tenant-admin"),
				Object:             NewEntityKey(TenantType, "restricted-tenant"),
				ExpectUnauthorized: true,
			},
			{
				Subject:       NewEntityKey(UserType, "ian"),
				Object:        NewEntityKey(TenantType, "tl-tenant"),
				ExpectActions: []Action{CanView},
			},
			{
				Subject:            NewEntityKey(UserType, "unknown"),
				Object:             NewEntityKey(TenantType, "tl-tenant"),
				ExpectUnauthorized: true,
			},
		}
		for _, tc := range checks {
			t.Run(tc.String(), func(t *testing.T) {
				ltk := dbTupleLookup(t, dbx, tc.TupleKey())
				srv := testServerWithUser(checker, tc)
				req, _ := http.NewRequest("GET", fmt.Sprintf("/tenants/%s", ltk.Object.Name), nil)
				rr := httptest.NewRecorder()
				srv.ServeHTTP(rr, req)
				checkHttpExpectError(t, tc, rr)
				gotActions := responseGetActions(t, rr.Body.Bytes())
				assert.ElementsMatch(t, tc.ExpectActions, gotActions)
			})
		}
	})

	// GROUPS
	t.Run("GroupList", func(t *testing.T) {
		checker := newTestChecker(t, fgaUrl, serverTestData)
		checks := []TestTuple{
			{
				Subject:    NewEntityKey(UserType, "tl-tenant-admin"),
				ExpectKeys: newEntityKeys(GroupType, "BA-group", "CT-group"),
			},
			{
				Subject:    NewEntityKey(UserType, "ian"),
				ExpectKeys: newEntityKeys(GroupType, "BA-group"),
			},
			{
				Subject:    NewEntityKey(UserType, "drew"),
				ExpectKeys: newEntityKeys(GroupType, "CT-group"),
			},
		}
		for _, tc := range checks {
			t.Run(tc.String(), func(t *testing.T) {
				srv := testServerWithUser(checker, tc)
				req, _ := http.NewRequest("GET", "/groups", nil)
				rr := httptest.NewRecorder()
				srv.ServeHTTP(rr, req)
				checkHttpExpectError(t, tc, rr)
				assert.ElementsMatch(
					t,
					ekGetNames(tc.ExpectKeys),
					responseGetNames(t, rr.Body.Bytes(), "groups", "name"),
				)
			})
		}
	})

	t.Run("GroupPermissions", func(t *testing.T) {
		checker := newTestChecker(t, fgaUrl, serverTestData)
		checks := []TestTuple{
			{
				Subject:       NewEntityKey(UserType, "tl-tenant-admin"),
				Object:        NewEntityKey(GroupType, "BA-group"),
				ExpectActions: []Action{CanView, CanEdit, CanEditMembers, CanCreateFeed, CanDeleteFeed},
			},
			{
				Subject:       NewEntityKey(UserType, "ian"),
				Object:        NewEntityKey(GroupType, "BA-group"),
				ExpectActions: []Action{CanView},
			},
			{
				Subject:       NewEntityKey(UserType, "drew"),
				Object:        NewEntityKey(GroupType, "CT-group"),
				ExpectActions: []Action{CanView, CanEdit},
			},
			{
				Subject:            NewEntityKey(UserType, "unknown"),
				Object:             NewEntityKey(GroupType, "CT-group"),
				ExpectUnauthorized: true,
			},
		}
		for _, tc := range checks {
			t.Run(tc.String(), func(t *testing.T) {
				ltk := dbTupleLookup(t, dbx, tc.TupleKey())
				srv := testServerWithUser(checker, tc)
				req, _ := http.NewRequest("GET", fmt.Sprintf("/groups/%s", ltk.Object.Name), nil)
				rr := httptest.NewRecorder()
				srv.ServeHTTP(rr, req)
				checkHttpExpectError(t, tc, rr)
				gotActions := responseGetActions(t, rr.Body.Bytes())
				assert.ElementsMatch(t, tc.ExpectActions, gotActions)
			})
		}
	})

	// FEEDS
	t.Run("FeedList", func(t *testing.T) {
		checker := newTestChecker(t, fgaUrl, serverTestData)
		checks := []TestTuple{
			{
				Subject:    NewEntityKey(UserType, "tl-tenant-admin"),
				ExpectKeys: newEntityKeys(TenantType, "BA", "CT"),
			},
			{
				Subject:    NewEntityKey(UserType, "ian"),
				ExpectKeys: newEntityKeys(TenantType, "BA"),
			},
			{
				Subject:    NewEntityKey(UserType, "drew"),
				ExpectKeys: newEntityKeys(TenantType, "CT"),
			},
			{
				Subject:    NewEntityKey(UserType, "unknown"),
				ExpectKeys: newEntityKeys(TenantType),
			},
		}
		for _, tc := range checks {
			t.Run(tc.String(), func(t *testing.T) {
				srv := testServerWithUser(checker, tc)
				req, _ := http.NewRequest("GET", "/feeds", nil)
				rr := httptest.NewRecorder()
				srv.ServeHTTP(rr, req)
				checkHttpExpectError(t, tc, rr)
				assert.ElementsMatch(
					t,
					ekGetNames(tc.ExpectKeys),
					responseGetNames(t, rr.Body.Bytes(), "feeds", "onestop_id"),
				)
			})
		}
	})

	t.Run("FeedPermissions", func(t *testing.T) {
		checker := newTestChecker(t, fgaUrl, serverTestData)
		checks := []TestTuple{
			{
				Subject:       NewEntityKey(UserType, "tl-tenant-admin"),
				Object:        NewEntityKey(FeedType, "BA"),
				ExpectActions: []Action{CanView, CanEdit, CanCreateFeedVersion, CanDeleteFeedVersion, CanSetGroup},
			},
			{
				Subject:       NewEntityKey(UserType, "ian"),
				Object:        NewEntityKey(FeedType, "BA"),
				ExpectActions: []Action{CanView},
			},
			{
				Subject:       NewEntityKey(UserType, "drew"),
				Object:        NewEntityKey(FeedType, "CT"),
				ExpectActions: []Action{CanView, CanEdit, CanCreateFeedVersion, CanDeleteFeedVersion},
			},
			{
				Subject:            NewEntityKey(UserType, "unknown"),
				Object:             NewEntityKey(FeedType, "CT"),
				ExpectUnauthorized: true,
			},
		}
		for _, tc := range checks {
			t.Run(tc.String(), func(t *testing.T) {
				ltk := dbTupleLookup(t, dbx, tc.TupleKey())
				srv := testServerWithUser(checker, tc)
				req, _ := http.NewRequest("GET", fmt.Sprintf("/feeds/%s", ltk.Object.Name), nil)
				rr := httptest.NewRecorder()
				srv.ServeHTTP(rr, req)
				checkHttpExpectError(t, tc, rr)
				gotActions := responseGetActions(t, rr.Body.Bytes())
				assert.ElementsMatch(t, tc.ExpectActions, gotActions)
			})
		}
	})

	// FEED VERSIONS
	t.Run("FeedVersionList", func(t *testing.T) {
		checker := newTestChecker(t, fgaUrl, serverTestData)
		checks := []TestTuple{
			{
				Subject:    NewEntityKey(UserType, "tl-tenant-admin"),
				ExpectKeys: newEntityKeys(FeedVersionType),
			},
			{
				Subject:    NewEntityKey(UserType, "ian"),
				ExpectKeys: newEntityKeys(FeedVersionType),
			},
			{
				Subject:    NewEntityKey(UserType, "drew"),
				ExpectKeys: newEntityKeys(FeedVersionType, "e535eb2b3b9ac3ef15d82c56575e914575e732e0"),
			},
		}
		for _, tc := range checks {
			t.Run(tc.String(), func(t *testing.T) {
				srv := testServerWithUser(checker, tc)
				req, _ := http.NewRequest("GET", "/feed_versions", nil)
				rr := httptest.NewRecorder()
				srv.ServeHTTP(rr, req)
				checkHttpExpectError(t, tc, rr)
				assert.ElementsMatch(
					t,
					ekGetNames(tc.ExpectKeys),
					responseGetNames(t, rr.Body.Bytes(), "feed_versions", "sha1"),
				)
			})
		}
	})

	t.Run("FeedVersionPermissions", func(t *testing.T) {
		checker := newTestChecker(t, fgaUrl, serverTestData)
		checks := []TestTuple{
			{
				Subject:       NewEntityKey(UserType, "tl-tenant-admin"),
				Object:        NewEntityKey(FeedVersionType, "e535eb2b3b9ac3ef15d82c56575e914575e732e0"),
				ExpectActions: []Action{CanView, CanEdit, CanEditMembers},
			},
			{
				Subject:       NewEntityKey(UserType, "ian"),
				Object:        NewEntityKey(FeedVersionType, "e535eb2b3b9ac3ef15d82c56575e914575e732e0"),
				ExpectActions: []Action{CanView},
			},
			{
				Subject:       NewEntityKey(UserType, "drew"),
				Object:        NewEntityKey(FeedVersionType, "e535eb2b3b9ac3ef15d82c56575e914575e732e0"),
				ExpectActions: []Action{CanView},
			},
			{
				Subject:            NewEntityKey(UserType, "unknown"),
				Object:             NewEntityKey(FeedVersionType, "e535eb2b3b9ac3ef15d82c56575e914575e732e0"),
				ExpectUnauthorized: true,
			},
		}
		for _, tc := range checks {
			t.Run(tc.String(), func(t *testing.T) {
				ltk := dbTupleLookup(t, dbx, tc.TupleKey())
				srv := testServerWithUser(checker, tc)
				req, _ := http.NewRequest("GET", fmt.Sprintf("/feed_versions/%s", ltk.Object.Name), nil)
				rr := httptest.NewRecorder()
				srv.ServeHTTP(rr, req)
				checkHttpExpectError(t, tc, rr)
				gotActions := responseGetActions(t, rr.Body.Bytes())
				assert.ElementsMatch(t, tc.ExpectActions, gotActions)
			})
		}
	})

}

func testServerWithUser(c *Checker, tk TestTuple) http.Handler {
	srv, _ := NewServer(c)
	srv = authn.UserDefaultMiddleware(stringOr(tk.CheckAsUser, tk.Subject.Name))(srv)
	return srv
}

func checkHttpExpectError(t testing.TB, tk TestTuple, rr *httptest.ResponseRecorder) {
	status := rr.Code
	if tk.ExpectUnauthorized {
		if status != http.StatusUnauthorized {
			t.Errorf("got error code %d, expected %d", status, http.StatusUnauthorized)
		}
	} else if tk.ExpectError {
		if status == http.StatusOK {
			t.Errorf("got status %d, expected non-200", status)
		}
	} else if status != http.StatusOK {
		t.Errorf("got error code %d, expected 200", status)
	}

}

func responseGetNames(t testing.TB, data []byte, path string, key string) []string {
	a := gjson.ParseBytes(data).Get(path)
	var ret []string
	for _, b := range a.Array() {
		ret = append(ret, b.Get(key).Str)
	}
	return ret
}

func ekGetNames(eks []EntityKey) []string {
	var ret []string
	for _, ek := range eks {
		ret = append(ret, ek.Name)
	}
	return ret
}

func responseGetActions(t testing.TB, data []byte) []Action {
	a := gjson.ParseBytes(data).Get("actions")
	var ret []Action
	for k, v := range a.Map() {
		if v.Bool() {
			a, err := azpb.ActionString(k)
			if err != nil {
				t.Errorf("invalid action %s", k)
			}
			ret = append(ret, a)
		}
	}
	return ret
}
