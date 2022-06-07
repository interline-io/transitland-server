package auth

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gorilla/mux"
	"github.com/stretchr/testify/assert"
)

func testAuthMiddleware(t *testing.T, req *http.Request, mwf mux.MiddlewareFunc, expectCode int, expectUser *User) {
	var user *User
	testHandler := func(w http.ResponseWriter, r *http.Request) {
		user = ForContext(r.Context())
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
		assert.Equal(t, user.Name, expectUser.Name)
		assert.Equal(t, user.IsAdmin, expectUser.IsAdmin)
		assert.Equal(t, user.IsAnon, expectUser.IsAnon)
		assert.Equal(t, user.IsUser, expectUser.IsUser)
	} else if expectUser == nil && user != nil {
		t.Errorf("got user, expected none")
	} else if expectUser != nil && user == nil {
		t.Errorf("got no user, expected user")
	}
}
func TestUserMiddleware(t *testing.T) {
	a := UserDefaultMiddleware()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	testAuthMiddleware(t, req, a, 200, &User{IsAnon: false, IsUser: true, IsAdmin: false})
}

func TestAdminMiddleware(t *testing.T) {
	a := AdminDefaultMiddleware()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	testAuthMiddleware(t, req, a, 200, &User{IsAnon: false, IsUser: true, IsAdmin: true})
}

func TestNoMiddleware(t *testing.T) {
	a, err := GetUserMiddleware("", AuthConfig{})
	if err != nil {
		t.Fatal(err)
	}
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	testAuthMiddleware(t, req, a, 200, nil)
}

func TestUserRequired(t *testing.T) {
	tcs := []struct {
		name string
		mwf  mux.MiddlewareFunc
		code int
		user *User
	}{
		{"with user", func(next http.Handler) http.Handler { return AdminDefaultMiddleware()(UserRequired(next)) }, 200, &User{IsAdmin: true, IsUser: true}},
		{"with user", func(next http.Handler) http.Handler { return UserDefaultMiddleware()(UserRequired(next)) }, 200, &User{IsUser: true}},
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
		mwf  mux.MiddlewareFunc
		code int
		user *User
	}{
		{"with admin", func(next http.Handler) http.Handler { return AdminDefaultMiddleware()(AdminRequired(next)) }, 200, &User{IsAdmin: true, IsUser: true}},
		{"with user", func(next http.Handler) http.Handler { return UserDefaultMiddleware()(AdminRequired(next)) }, 401, nil}, // mw kills request before handler
		{"no user", func(next http.Handler) http.Handler { return AdminRequired(next) }, 401, nil},
	}
	for _, tc := range tcs {
		t.Run(tc.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/", nil)
			testAuthMiddleware(t, req, tc.mwf, tc.code, tc.user)
		})
	}
}
