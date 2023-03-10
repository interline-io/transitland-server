package auth

import (
	"context"
	"net/http"

	"github.com/go-redis/redis/v8"
)

type MiddlewareFunc func(http.Handler) http.Handler

type AuthConfig struct {
	DefaultUsername      string
	GatekeeperEndpoint   string
	GatekeeperParam      string
	GatekeeperSelector   string
	GatekeeperAllowError bool
	JwtAudience          string
	JwtIssuer            string
	JwtPublicKeyFile     string
	UserHeader           string
}

// GetUserMiddleware returns a middleware that sets user details.
func GetUserMiddleware(authType string, cfg AuthConfig, client *redis.Client) (MiddlewareFunc, error) {
	// Setup auth; default is all users will be anonymous.
	switch authType {
	case "admin":
		return AdminDefaultMiddleware(cfg.DefaultUsername), nil
	case "user":
		return UserDefaultMiddleware(cfg.DefaultUsername), nil
	case "jwt":
		return JWTMiddleware(cfg.JwtAudience, cfg.JwtIssuer, cfg.JwtPublicKeyFile)
	case "header":
		return UserHeaderMiddleware(cfg.UserHeader)
	case "kong":
		return UserHeaderMiddleware("x-consumer-username")
	case "gatekeeper":
		return GatekeeperMiddleware(client, cfg.GatekeeperEndpoint, cfg.GatekeeperParam, cfg.GatekeeperSelector, cfg.GatekeeperAllowError)
	}
	return func(next http.Handler) http.Handler {
		return next
	}, nil
}

// AdminDefaultMiddleware uses a default "admin" context.
func AdminDefaultMiddleware(defaultName string) func(http.Handler) http.Handler {
	return NewUserDefaultMiddleware(func() *User { return NewUser(defaultName).WithRoles("user", "admin") })
}

// UserDefaultMiddleware uses a default "user" context.
func UserDefaultMiddleware(defaultName string) func(http.Handler) http.Handler {
	return NewUserDefaultMiddleware(func() *User { return NewUser(defaultName).WithRoles("user") })
}

// NewUserDefaultMiddleware uses a default "user" context.
func NewUserDefaultMiddleware(cb func() *User) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			user := cb()
			r = r.WithContext(context.WithValue(r.Context(), userCtxKey, user))
			next.ServeHTTP(w, r)
		})
	}
}

// AdminRequired limits a request to admin privileges.
func AdminRequired(next http.Handler) http.Handler {
	return RoleRequired("admin")(next)
}

// UserRequired limits a request to user privileges.
func UserRequired(next http.Handler) http.Handler {
	return RoleRequired("user")(next)
}

func RoleRequired(role string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx := r.Context()
			user := ForContext(ctx)
			if user == nil || !user.HasRole(role) {
				http.Error(w, `{"error":"permission denied"}`, http.StatusUnauthorized)
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}
