package auth

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
	"time"

	"github.com/go-redis/redismock/v8"
	"github.com/gorilla/mux"
	"github.com/interline-io/transitland-server/internal/ecache"
	"github.com/stretchr/testify/assert"
)

func TestGatekeeper(t *testing.T) {
	// Mock users
	testEmail := "test@transit.land"
	testRole := "test_role"
	users := map[string]*User{
		testEmail:              NewUser(testEmail).WithRoles("user", testRole),
		"refresh@transit.land": NewUser("refresh@transit.land").WithRoles("user", "refresh_test"),
	}

	// Mock gatekeeper api interface
	requestCounts := map[string]int{}
	requestCountLock := sync.Mutex{}
	ts200 := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Log("gatekeeper mock api get:", r.URL)
		u := r.URL.Query()
		var user *User
		if a := u["user"]; len(a) > 0 {
			user = users[a[0]]
		}
		if user != nil {
			requestCountLock.Lock()
			requestCounts[user.Name] += 1
			requestCountLock.Unlock()
			umap := map[string]any{
				"name":  user.Name,
				"roles": user.roles,
			}
			jb, err := json.Marshal(umap)
			if err != nil {
				panic(err)
			}
			w.WriteHeader(200)
			w.Write(jb)
			return
		}
		http.Error(w, "error", 404)
	}))
	defer ts200.Close()

	// Invalid response with error
	tsInvalid := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Log("gatekeeper mock invalid get:", r.URL)
		http.Error(w, "error", 500)
	}))
	defer tsInvalid.Close()

	// Invalid response with 200 OK
	tsInvalidOk := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Log("gatekeeper mock invalid 200 get:", r.URL)
		w.WriteHeader(200)
		w.Write([]byte("not json data"))
	}))
	defer tsInvalidOk.Close()

	// Close after 10 seconds with no response
	tsTimeout := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Log("gatekeeper mock timeout get:", r.URL)
		ctx := r.Context()
		select {
		case <-ctx.Done():
			t.Log("gatekeeper mock timeout: client closed connection")
		case <-time.After(10 * time.Second):
			t.Log("gatekeeper mock timeout: closing after timeout")
		}
	}))
	defer tsTimeout.Close()

	// Middleware create helpers
	gkHelper := func(next http.Handler, em string, endpoint string, param string, roleKey string, allowError bool) http.Handler {
		// Does not start update process
		gk := NewGatekeeper(nil, endpoint, param, roleKey)
		return UserDefaultMiddleware(em)(newGatekeeperMiddleware(gk, allowError)(next))
	}

	// Tests
	testStartTime := time.Now()
	tcs := []struct {
		name  string
		mwf   mux.MiddlewareFunc
		code  int
		user  *User
		after func(*testing.T)
	}{
		{
			"endpoint returns specified roles",
			func(next http.Handler) http.Handler {
				return gkHelper(next, testEmail, ts200.URL, "user", "roles", false)
			},
			200,
			NewUser(testEmail).WithRoles("user", testRole),
			nil,
		},
		{
			"unknown user returns 401",
			func(next http.Handler) http.Handler {
				return gkHelper(next, "unknown@transit.land", ts200.URL, "user", "roles", false)
			},
			401,
			nil,
			nil,
		},
		{
			"locally cached value ok when endpoint available",
			func(next http.Handler) http.Handler {
				gk := NewGatekeeper(nil, ts200.URL, "user", "roles")
				gk.cache.SetTTL(nil, testEmail, []string{"user", testRole}, 0, 0)
				return UserDefaultMiddleware(testEmail)(newGatekeeperMiddleware(gk, false)(next))
			},
			200,
			NewUser(testEmail).WithRoles("user", testRole),
			nil,
		},
		{
			"locally cached value ok when endpoint down",
			func(next http.Handler) http.Handler {
				gk := NewGatekeeper(nil, tsTimeout.URL, "user", "roles")
				gk.RequestTimeout = 100 * time.Millisecond
				gk.cache.SetTTL(nil, testEmail, []string{"user", testRole}, 0, 0)
				return UserDefaultMiddleware(testEmail)(newGatekeeperMiddleware(gk, false)(next))
			},
			200,
			NewUser(testEmail).WithRoles("user", testRole),
			nil,
		},
		{
			"unknown user skipped when allowFail = true",
			func(next http.Handler) http.Handler {
				return gkHelper(next, "other@transit.land", ts200.URL, "user", "roles", true)
			},
			200,
			NewUser("other@transit.land").WithRoles("user"),
			nil,
		},
		{
			"invalid data returns 401",
			func(next http.Handler) http.Handler {
				return gkHelper(next, testEmail, tsInvalidOk.URL, "user", "roles", false)
			},
			401,
			nil,
			nil,
		},
		{
			"invalid data allowed when allowFail = true",
			func(next http.Handler) http.Handler {
				return gkHelper(next, testEmail, tsInvalidOk.URL, "user", "roles", true)
			},
			200,
			NewUser(testEmail).WithRoles("user"),
			nil,
		},
		{
			"invalid response allowed when allowFail = true",
			func(next http.Handler) http.Handler {
				return gkHelper(next, "other@transit.land", tsInvalid.URL, "user", "roles", true)
			},
			200,
			NewUser("other@transit.land").WithRoles("user"),
			nil,
		},
		{
			"invalid response returns 401",
			func(next http.Handler) http.Handler {
				return gkHelper(next, "other@transit.land", tsInvalid.URL, "user", "roles", false)
			},
			401,
			nil,
			nil,
		},
		{
			"endpoint down, returns 401 after request timeout",
			func(next http.Handler) http.Handler {
				testStartTime = time.Now()
				gk := NewGatekeeper(nil, tsTimeout.URL, "user", "roles")
				gk.RequestTimeout = 100 * time.Millisecond
				return UserDefaultMiddleware(testEmail)(newGatekeeperMiddleware(gk, false)(next))
			},
			401,
			nil,
			func(t *testing.T) {
				// Check at least 100ms have elapsed
				elapsedTime := time.Now().UnixNano() - testStartTime.UnixNano()
				assert.GreaterOrEqual(t, elapsedTime, int64(100*1e6)) // 100*1e6 = 100ms in nanoseconds
				fmt.Println("elapsedTime:", elapsedTime)
			},
		},
		{
			"endpoint down, ignored when allowFail = true",
			func(next http.Handler) http.Handler {
				gk := NewGatekeeper(nil, tsTimeout.URL, "user", "roles")
				gk.RequestTimeout = 100 * time.Millisecond
				return UserDefaultMiddleware(testEmail)(newGatekeeperMiddleware(gk, true)(next))
			},
			200,
			NewUser(testEmail).WithRoles("user"),
			nil,
		},
		{
			"redis cached value ok when endpoint down",
			func(next http.Handler) http.Handler {
				db, mock := redismock.NewClientMock()
				mock.ExpectGet(cacheRedisKey("gatekeeper", testEmail)).SetVal(cacheItemJson([]string{testRole}, 0))
				gk := NewGatekeeper(db, tsTimeout.URL, "user", "roles")
				gk.RequestTimeout = 100 * time.Millisecond
				return UserDefaultMiddleware(testEmail)(newGatekeeperMiddleware(gk, false)(next))
			},
			200,
			NewUser(testEmail).WithRoles("user", testRole),
			nil,
		},
		{
			"gatekeeper refreshes contexts in background",
			func(next http.Handler) http.Handler {
				gk := NewGatekeeper(nil, ts200.URL, "user", "roles")
				gk.recheckTtl = 1 * time.Millisecond
				gk.Start(10 * time.Millisecond)
				return UserDefaultMiddleware("refresh@transit.land")(newGatekeeperMiddleware(gk, false)(next))
			},
			200,
			NewUser("refresh@transit.land").WithRoles("user", "refresh_test"),
			func(t *testing.T) {
				// Request count should be at least 10
				time.Sleep(100 * time.Millisecond)
				requestCountLock.Lock()
				a := requestCounts["refresh@transit.land"]
				requestCountLock.Unlock()
				assert.GreaterOrEqual(t, a, 10)
			},
		},
	}
	for _, tc := range tcs {
		t.Run(tc.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/", nil)
			testAuthMiddleware(t, req, tc.mwf, tc.code, tc.user)
			if tc.after != nil {
				tc.after(t)
			}
		})
	}
}

// Must be the same as ecache
func cacheItemJson(roles []string, ttl time.Duration) string {
	a := ecache.Item[[]string]{Value: roles, ExpiresAt: time.Now().Add(ttl), RecheckAt: time.Now().Add(ttl)}
	b, err := json.Marshal(&a)
	if err != nil {
		panic(err)
	}
	return string(b)
}

// This must be the same as ecache.redisKey
func cacheRedisKey(topic string, key string) string {
	return fmt.Sprintf("ecache:%s:%s", topic, key)
}
