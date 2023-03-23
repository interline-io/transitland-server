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
	"github.com/interline-io/transitland-server/internal/ecache"
	"github.com/stretchr/testify/assert"
)

func TestGatekeeper(t *testing.T) {
	// Mock users
	testEmail := "test@transit.land"
	testRole := "test_role"

	// Test server
	gkts := GatekeeperTestServer{}
	gkts.AddUser(testEmail, newCtxUser(testEmail).WithRoles(testRole))
	gkts.AddUser("refresh@transit.land", newCtxUser("refresh@transit.land").WithRoles("refresh_test"))

	// Mock gatekeeper api interface
	ts200 := httptest.NewServer(&gkts)
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
	gkHelper := func(next http.Handler, em string, endpoint string, param string, roleKey string, eidKey string, allowError bool) http.Handler {
		// Does not start update process
		gk := NewGatekeeper(nil, endpoint, param, roleKey, eidKey)
		return UserDefaultMiddleware(em)(newGatekeeperMiddleware(gk, allowError)(next))
	}

	// Tests
	testStartTime := time.Now()
	tcs := []struct {
		name  string
		mwf   MiddlewareFunc
		code  int
		user  User
		after func(*testing.T)
	}{
		{
			"endpoint returns specified roles",
			func(next http.Handler) http.Handler {
				return gkHelper(next, testEmail, ts200.URL, "user", "roles", "external_ids", false)
			},
			200,
			newCtxUser(testEmail).WithRoles(testRole),
			nil,
		},
		{
			"unknown user returns 401",
			func(next http.Handler) http.Handler {
				return gkHelper(next, "unknown@transit.land", ts200.URL, "user", "roles", "external_ids", false)
			},
			401,
			nil,
			nil,
		},
		{
			"locally cached value ok when endpoint available",
			func(next http.Handler) http.Handler {
				u := gkCacheItem{Name: testEmail, Roles: []string{testRole}}
				gk := NewGatekeeper(nil, ts200.URL, "user", "roles", "external_ids")
				gk.cache.SetTTL(nil, testEmail, u, 0, 0)
				return UserDefaultMiddleware(testEmail)(newGatekeeperMiddleware(gk, false)(next))
			},
			200,
			newCtxUser(testEmail).WithRoles(testRole),
			nil,
		},
		{
			"locally cached value ok when endpoint down",
			func(next http.Handler) http.Handler {
				u := gkCacheItem{Name: testEmail, Roles: []string{testRole}}
				gk := NewGatekeeper(nil, tsTimeout.URL, "user", "roles", "external_ids")
				gk.RequestTimeout = 100 * time.Millisecond
				gk.cache.SetTTL(nil, testEmail, u, 0, 0)
				return UserDefaultMiddleware(testEmail)(newGatekeeperMiddleware(gk, false)(next))
			},
			200,
			newCtxUser(testEmail).WithRoles(testRole),
			nil,
		},
		{
			"unknown user skipped when allowFail = true",
			func(next http.Handler) http.Handler {
				return gkHelper(next, "other@transit.land", ts200.URL, "user", "roles", "external_ids", true)
			},
			200,
			newCtxUser("other@transit.land"),
			nil,
		},
		{
			"invalid data returns 401",
			func(next http.Handler) http.Handler {
				return gkHelper(next, testEmail, tsInvalidOk.URL, "user", "roles", "external_ids", false)
			},
			401,
			nil,
			nil,
		},
		{
			"invalid data allowed when allowFail = true",
			func(next http.Handler) http.Handler {
				return gkHelper(next, testEmail, tsInvalidOk.URL, "user", "roles", "external_ids", true)
			},
			200,
			newCtxUser(testEmail),
			nil,
		},
		{
			"invalid response allowed when allowFail = true",
			func(next http.Handler) http.Handler {
				return gkHelper(next, "other@transit.land", tsInvalid.URL, "user", "roles", "external_ids", true)
			},
			200,
			newCtxUser("other@transit.land"),
			nil,
		},
		{
			"invalid response returns 401",
			func(next http.Handler) http.Handler {
				return gkHelper(next, "other@transit.land", tsInvalid.URL, "user", "roles", "external_ids", false)
			},
			401,
			nil,
			nil,
		},
		{
			"endpoint down, returns 401 after request timeout",
			func(next http.Handler) http.Handler {
				testStartTime = time.Now()
				gk := NewGatekeeper(nil, tsTimeout.URL, "user", "roles", "external_ids")
				gk.RequestTimeout = 100 * time.Millisecond
				return UserDefaultMiddleware(testEmail)(newGatekeeperMiddleware(gk, false)(next))
			},
			401,
			nil,
			func(t *testing.T) {
				// Check at least 100ms have elapsed
				elapsedTime := time.Now().UnixNano() - testStartTime.UnixNano()
				assert.GreaterOrEqual(t, elapsedTime, int64(100*1e6)) // 100*1e6 = 100ms in nanoseconds
				t.Log("elapsedTime:", elapsedTime)
			},
		},
		{
			"endpoint down, ignored when allowFail = true",
			func(next http.Handler) http.Handler {
				gk := NewGatekeeper(nil, tsTimeout.URL, "user", "roles", "external_ids")
				gk.RequestTimeout = 100 * time.Millisecond
				return UserDefaultMiddleware(testEmail)(newGatekeeperMiddleware(gk, true)(next))
			},
			200,
			newCtxUser(testEmail),
			nil,
		},
		{
			"redis cached value ok when endpoint down",
			func(next http.Handler) http.Handler {
				u := gkCacheItem{Name: testEmail, Roles: []string{testRole}}
				db, mock := redismock.NewClientMock()
				mock.ExpectGet(cacheRedisKey("gatekeeper", testEmail)).SetVal(cacheItemJson(u, 0))
				gk := NewGatekeeper(db, tsTimeout.URL, "user", "roles", "external_ids")
				gk.RequestTimeout = 100 * time.Millisecond
				return UserDefaultMiddleware(testEmail)(newGatekeeperMiddleware(gk, false)(next))
			},
			200,
			newCtxUser(testEmail).WithRoles(testRole),
			nil,
		},
		{
			"gatekeeper refreshes contexts in background",
			func(next http.Handler) http.Handler {
				gk := NewGatekeeper(nil, ts200.URL, "user", "roles", "external_ids")
				gk.recheckTtl = 1 * time.Millisecond
				gk.Start(10 * time.Millisecond)
				return UserDefaultMiddleware("refresh@transit.land")(newGatekeeperMiddleware(gk, false)(next))
			},
			200,
			newCtxUser("refresh@transit.land").WithRoles("refresh_test"),
			func(t *testing.T) {
				// Request count should be at least 10
				time.Sleep(100 * time.Millisecond)
				a := gkts.counts["refresh@transit.land"]
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
func cacheItemJson(user gkCacheItem, ttl time.Duration) string {
	a := ecache.Item[gkCacheItem]{Value: user, ExpiresAt: time.Now().Add(ttl), RecheckAt: time.Now().Add(ttl)}
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

//////////

// Trivial implementation of Gatekeeper for testing purposes
type GatekeeperTestServer struct {
	users  map[string]User
	counts map[string]int
	lock   sync.Mutex
}

func (gk *GatekeeperTestServer) AddUser(key string, user User) {
	gk.lock.Lock()
	defer gk.lock.Unlock()
	if gk.users == nil {
		gk.users = map[string]User{}
	}
	gk.users[key] = newCtxUser(user.Name()).WithRoles(user.Roles()...)

}

func (gk *GatekeeperTestServer) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	gk.lock.Lock()
	defer gk.lock.Unlock()
	u := r.URL.Query()
	var user User
	if a := u["user"]; len(a) > 0 {
		user = gk.users[a[0]]
	}
	if user != nil {
		if gk.counts == nil {
			gk.counts = map[string]int{}
		}
		gk.counts[user.Name()] += 1
		umap := map[string]any{
			"name":  user.Name(),
			"roles": user.Roles(),
		}
		jb, err := json.Marshal(umap)
		if err != nil {
			http.Error(w, "json error", 500)
		}
		w.WriteHeader(200)
		w.Write(jb)
		return
	}
	http.Error(w, "error", 404)
}
