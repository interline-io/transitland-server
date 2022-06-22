package auth

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"

	"github.com/gorilla/mux"
	"github.com/rs/zerolog/log"
	"github.com/tidwall/gjson"
)

// Gatekeeper checks an external endpoint for a list of roles
func Gatekeeper(endpoint string, roleKey string) (mux.MiddlewareFunc, error) {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx := r.Context()
			user := ForContext(ctx)
			if user != nil {
				log.Trace().Str("user", user.Name).Msg("checking gatekeeper")
				checkedUser, err := getGatekeeperUser(ctx, endpoint, user.Name, roleKey)
				if err != nil {
					log.Trace().Err(err).Str("user", user.Name).Msg("gatekeeper error")
					http.Error(w, "error", http.StatusInternalServerError)
					return
				}
				if checkedUser != nil {
					user.Roles = append(user.Roles, checkedUser.Roles...)
				}
				r = r.WithContext(context.WithValue(r.Context(), userCtxKey, user))
			}
			next.ServeHTTP(w, r)
		})
	}, nil
}

func getGatekeeperUser(ctx context.Context, endpoint string, email string, roleKey string) (*User, error) {
	u, _ := url.Parse(endpoint)
	rq := u.Query()
	rq.Set("email", email)
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
	user := User{}
	for _, r := range gjson.Get(string(body), roleKey).Array() {
		user.Roles = append(user.Roles, r.String())
	}
	return &user, nil
}
