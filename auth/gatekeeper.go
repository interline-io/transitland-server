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
	"github.com/gorilla/mux"
	"github.com/interline-io/transitland-lib/log"
	"github.com/interline-io/transitland-server/internal/ecache"
	"github.com/tidwall/gjson"
)

// GatekeeperMiddleware checks an external endpoint for a list of roles
func GatekeeperMiddleware(client *redis.Client, endpoint string, param string, roleKey string) (mux.MiddlewareFunc, error) {
	gk := NewGatekeeper(client, endpoint, param, roleKey)
	gk.Start()
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx := r.Context()
			user := ForContext(ctx)
			if user != nil && user.Name != "" {
				checkedRoles, err := gk.GetUser(ctx, user.Name)
				if err != nil {
					log.Trace().Err(err).Str("user", user.Name).Msg("gatekeeper error")
					http.Error(w, "error", http.StatusInternalServerError)
					return
				}
				log.Trace().Str("user", user.Name).Strs("roles", checkedRoles).Msg("gatekeeper roles")
				user.AddRoles(checkedRoles...)
				r = r.WithContext(context.WithValue(r.Context(), userCtxKey, user))
			}
			next.ServeHTTP(w, r)
		})
	}, nil
}

type Gatekeeper struct {
	endpoint string
	roleKey  string
	param    string
	ttl      time.Duration
	cache    *ecache.Cache[[]string]
}

func NewGatekeeper(client *redis.Client, endpoint string, param string, roleKey string) *Gatekeeper {
	return &Gatekeeper{
		endpoint: endpoint,
		roleKey:  roleKey,
		param:    param,
		ttl:      5 * 60 * time.Second,
		cache:    ecache.NewCache[[]string](client, "gatekeeper"),
	}
}

func (gk *Gatekeeper) GetUser(ctx context.Context, userKey string) ([]string, error) {
	roles, ok := gk.cache.Get(userKey)
	if !ok {
		var err error
		roles, err = gk.getUser(ctx, userKey)
		if err != nil {
			return nil, err
		}
		gk.cache.SetTTL(userKey, roles, gk.ttl, 24*time.Hour)
	}
	return roles, nil
}

func (gk *Gatekeeper) Start() {
	ticker := time.NewTicker(1 * time.Second)
	go func() {
		for t := range ticker.C {
			_ = t
			gk.updateUsers(context.TODO())
		}
	}()
}

func (gk *Gatekeeper) updateUsers(ctx context.Context) {
	// This can be improved to avoid races
	keys := gk.cache.GetRecheckKeys()
	for _, k := range keys {
		if roles, err := gk.getUser(ctx, k); err != nil {
			// Failed :(
			// Log but do not update cached value
		} else {
			gk.cache.SetTTL(k, roles, gk.ttl, 24*time.Hour)
		}
	}
}

func (gk *Gatekeeper) getUser(ctx context.Context, userKey string) ([]string, error) {
	u, _ := url.Parse(gk.endpoint)
	rq := u.Query()
	rq.Set(gk.param, userKey)
	u.RawQuery = rq.Encode()
	rctx, cf := context.WithTimeout(ctx, 1*time.Second)
	defer cf()
	req, err := http.NewRequestWithContext(rctx, "GET", u.String(), nil)
	if err != nil {
		return nil, errors.New("invalid request")
	}
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 400 {
		return nil, fmt.Errorf("response status code: %d", resp.StatusCode)
	}
	body, err := io.ReadAll(resp.Body)
	var roles []string
	for _, r := range gjson.Get(string(body), gk.roleKey).Array() {
		roles = append(roles, r.String())
	}
	return roles, nil
}
