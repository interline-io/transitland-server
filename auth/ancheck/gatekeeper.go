package ancheck

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
	"github.com/interline-io/transitland-server/auth/authn"
	"github.com/interline-io/transitland-server/internal/ecache"
	"github.com/interline-io/transitland-server/internal/util"
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
			if user := authn.ForContext(ctx); user != nil && user.ID() != "" {
				checkUser, err := gk.GetUser(ctx, user.ID())
				if err != nil {
					log.Error().Err(err).Msg("gatekeeper error")
					if !allowError {
						http.Error(w, util.MakeJsonError(http.StatusText(http.StatusUnauthorized)), http.StatusUnauthorized)
						return
					}
				} else if checkUser.ID() != "" {
					r = r.WithContext(authn.WithUser(r.Context(), checkUser))
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

func (gk *Gatekeeper) GetUser(ctx context.Context, userKey string) (authn.User, error) {
	gkUser, ok := gk.cache.Get(ctx, userKey)
	if !ok {
		var err error
		gkUser, err = gk.updateUser(ctx, userKey)
		if err != nil {
			return nil, err
		}
	}
	user := authn.NewCtxUser(gkUser.ID, "", "").WithRoles(gkUser.Roles...).WithExternalData(gkUser.ExternalData)
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
	for _, userKey := range keys {
		gk.updateUser(ctx, userKey)
	}
}

func (gk *Gatekeeper) updateUser(ctx context.Context, userKey string) (gkCacheItem, error) {
	gkUser, err := gk.requestUser(ctx, userKey)
	if err != nil {
		log.Error().Err(err).Str("user", userKey).Msg("gatekeeper requestUser failed")
		return gkUser, err
	}
	log.Trace().Str("user", userKey).Strs("roles", gkUser.Roles).Any("external_data", gkUser.ExternalData).Msg("gatekeeper requestUser ok")
	gk.cache.SetTTL(ctx, userKey, gkUser, gk.recheckTtl, 24*time.Hour)
	return gkUser, nil
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
	if err != nil {
		return gkCacheItem{}, err
	}
	if !gjson.Valid(string(body)) {
		return gkCacheItem{}, errors.New("invalid json")
	}
	parsed := gjson.ParseBytes(body)

	// Process roles and external IDs
	item := gkCacheItem{
		ID:           userKey,
		Roles:        []string{},
		ExternalData: map[string]string{},
	}
	item.ExternalData["gatekeeper"] = string(body)
	for _, r := range parsed.Get(gk.roleKey).Array() {
		item.Roles = append(item.Roles, r.String())
		// TODO: temporarily map "tl_admin" role to "admin" role.
		if r.String() == "tl_admin" {
			item.Roles = append(item.Roles, "admin")
		}
	}
	for k, v := range parsed.Get(gk.eidKey).Map() {
		item.ExternalData[k] = v.String()
	}
	return item, nil
}

// gkCacheItem needed for internal cached representation of ctxUser (Roles/ExternalData as exported fields)
type gkCacheItem struct {
	ID           string
	Roles        []string
	ExternalData map[string]string
}
