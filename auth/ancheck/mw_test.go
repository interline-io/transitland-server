package ancheck

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/interline-io/transitland-server/auth/authn"
	"github.com/stretchr/testify/assert"
)

func newCtxUser(id string) authn.CtxUser {
	return authn.NewCtxUser(id, "", "")
}

type userWithRoles interface {
	authn.User
	Roles() []string
}

func TestUserMiddleware(t *testing.T) {
	a := UserDefaultMiddleware("test")
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	testAuthMiddleware(t, req, a, 200, authn.NewCtxUser("test", "", ""))
}

func TestAdminMiddleware(t *testing.T) {
	a := AdminDefaultMiddleware("test")
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	testAuthMiddleware(t, req, a, 200, authn.NewCtxUser("test", "", "").WithRoles("admin"))
}

func TestNoMiddleware(t *testing.T) {
	a, err := GetUserMiddleware("", AuthConfig{}, nil)
	if err != nil {
		t.Fatal(err)
	}
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	testAuthMiddleware(t, req, a, 200, nil)
}

func testAuthMiddleware(t *testing.T, req *http.Request, mwf MiddlewareFunc, expectCode int, expectUser userWithRoles) {
	var user authn.User
	testHandler := func(w http.ResponseWriter, r *http.Request) {
		user = authn.ForContext(r.Context())
	}
	router := http.NewServeMux()
	router.HandleFunc("/", testHandler)
	//
	a := mwf(router)
	w := httptest.NewRecorder()
	a.ServeHTTP(w, req)
	//
	assert.Equal(t, expectCode, w.Result().StatusCode)
	if expectUser != nil && user != nil {
		assert.Equal(t, expectUser.ID(), user.ID())
		for _, checkRole := range expectUser.Roles() {
			assert.Equalf(t, true, user.HasRole(checkRole), "checking role '%s'", checkRole)
		}
	} else if expectUser == nil && user != nil {
		t.Errorf("got user, expected none")
	} else if expectUser != nil && user == nil {
		t.Errorf("got no user, expected user")
	}
}

func TestUserRequired(t *testing.T) {
	tcs := []struct {
		name string
		mwf  MiddlewareFunc
		code int
		user userWithRoles
	}{
		{"with user", func(next http.Handler) http.Handler { return AdminDefaultMiddleware("test")(UserRequired(next)) }, 200, newCtxUser("test").WithRoles("admin")},
		{"with user", func(next http.Handler) http.Handler { return UserDefaultMiddleware("test")(UserRequired(next)) }, 200, newCtxUser("test")},
		{"no user", func(next http.Handler) http.Handler { return UserRequired(next) }, 401, nil},
	}
	for _, tc := range tcs {
		t.Run(tc.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/", nil)
			testAuthMiddleware(t, req, tc.mwf, tc.code, tc.user)
		})
	}
}

func TestAdminRequired(t *testing.T) {
	tcs := []struct {
		name string
		mwf  MiddlewareFunc
		code int
		user userWithRoles
	}{
		{"with admin", func(next http.Handler) http.Handler { return AdminDefaultMiddleware("test")(AdminRequired(next)) }, 200, newCtxUser("test").WithRoles("admin")},
		{"with user", func(next http.Handler) http.Handler { return UserDefaultMiddleware("test")(AdminRequired(next)) }, 401, nil}, // mw kills request before handler
		{"no user", func(next http.Handler) http.Handler { return AdminRequired(next) }, 401, nil},
	}
	for _, tc := range tcs {
		t.Run(tc.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/", nil)
			testAuthMiddleware(t, req, tc.mwf, tc.code, tc.user)
		})
	}
}
