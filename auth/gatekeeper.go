package auth

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/interline-io/transitland-lib/log"
	"github.com/interline-io/transitland-server/internal/ecache"
	"github.com/tidwall/gjson"
)

// GatekeeperMiddleware checks an external endpoint for a list of roles
func GatekeeperMiddleware(client *redis.Client, endpoint string, param string, roleKey string, eidKey string, allowError bool) (MiddlewareFunc, error) {
	gk := NewGatekeeper(client, endpoint, param, roleKey, eidKey)
	gk.Start(60 * time.Second)
	return newGatekeeperMiddleware(gk, allowError), nil
}

func newGatekeeperMiddleware(gk *Gatekeeper, allowError bool) MiddlewareFunc {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Check context for a user name; if it is present, replace user context with gatekeeper user
			ctx := r.Context()
			if user := ForContext(ctx); user != nil && user.Name() != "" {
				checkUser, err := gk.GetUser(ctx, user.Name())
				if err != nil {
					log.Error().Err(err).Str("user", user.Name()).Msg("gatekeeper error")
					if !allowError {
						http.Error(w, "error", http.StatusUnauthorized)
						return
					}
				} else {
					log.Trace().Str("user", checkUser.Name()).Strs("roles", checkUser.Roles()).Msg("gatekeeper roles")
					r = r.WithContext(context.WithValue(r.Context(), userCtxKey, checkUser))
				}
			}
			next.ServeHTTP(w, r)
		})
	}
}

type Gatekeeper struct {
	RequestTimeout time.Duration
	endpoint       string
	roleKey        string
	eidKey         string
	param          string
	recheckTtl     time.Duration
	cache          *ecache.Cache[gkCacheItem]
}

func NewGatekeeper(client *redis.Client, endpoint string, param string, roleKey string, eidKey string) *Gatekeeper {
	gk := &Gatekeeper{
		RequestTimeout: 1 * time.Second,
		endpoint:       endpoint,
		roleKey:        roleKey,
		eidKey:         eidKey,
		param:          param,
		recheckTtl:     5 * 60 * time.Second,
		cache:          ecache.NewCache[gkCacheItem](client, "gatekeeper"),
	}
	return gk
}

func (gk *Gatekeeper) GetUserRoles(ctx context.Context, userKey string) ([]string, error) {
	u, err := gk.GetUser(ctx, userKey)
	if err != nil {
		return nil, err
	}
	return u.Roles(), nil
}

func (gk *Gatekeeper) GetUser(ctx context.Context, userKey string) (User, error) {
	gkUser, ok := gk.cache.Get(ctx, userKey)
	if !ok {
		var err error
		gkUser, err = gk.requestUser(ctx, userKey)
		if err != nil {
			return nil, err
		}
		gk.cache.SetTTL(ctx, userKey, gkUser, gk.recheckTtl, 24*time.Hour)
	}
	user := newCtxUser(gkUser.Name).WithRoles(gkUser.Roles...)
	return user, nil
}

func (gk *Gatekeeper) Start(t time.Duration) {
	ticker := time.NewTicker(t)
	go func() {
		for t := range ticker.C {
			_ = t
			gk.updateUsers(context.Background())
		}
	}()
}

func (gk *Gatekeeper) updateUsers(ctx context.Context) {
	// This can be improved to avoid races
	keys := gk.cache.GetRecheckKeys(ctx)
	for _, k := range keys {
		if gkUser, err := gk.requestUser(ctx, k); err != nil {
			// Failed :(
			// Log but do not update cached value
		} else {
			gk.cache.SetTTL(ctx, k, gkUser, gk.recheckTtl, 24*time.Hour)
		}
	}
}

func (gk *Gatekeeper) requestUser(ctx context.Context, userKey string) (gkCacheItem, error) {
	u, _ := url.Parse(gk.endpoint)
	rq := u.Query()
	rq.Set(gk.param, userKey)
	u.RawQuery = rq.Encode()
	rctx, cf := context.WithTimeout(ctx, gk.RequestTimeout)
	defer cf()
	req, err := http.NewRequestWithContext(rctx, "GET", u.String(), nil)
	if err != nil {
		return gkCacheItem{}, errors.New("invalid request")
	}
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return gkCacheItem{}, err
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 400 {
		return gkCacheItem{}, fmt.Errorf("response status code: %d", resp.StatusCode)
	}
	// Read response
	body, err := io.ReadAll(resp.Body)
	if !gjson.Valid(string(body)) {
		return gkCacheItem{}, errors.New("invalid json")
	}
	parsed := gjson.ParseBytes(body)

	// Process roles and external IDs
	item := gkCacheItem{
		Name:        userKey,
		Roles:       []string{},
		ExternalIDs: map[string]string{},
	}
	for _, r := range parsed.Get(gk.roleKey).Array() {
		item.Roles = append(item.Roles, r.String())
	}
	for k, v := range parsed.Get(gk.eidKey).Map() {
		item.ExternalIDs[k] = v.String()
	}
	return item, nil
}

// gkCacheItem needed for internal cached representation of ctxUser (Roles/ExternalIDs as exported fields)
type gkCacheItem struct {
	Name        string
	Roles       []string
	ExternalIDs map[string]string
}
