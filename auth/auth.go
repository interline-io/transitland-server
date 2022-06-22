package auth

import (
	"context"
	"net/http"

	"github.com/gorilla/mux"
)

type AuthConfig struct {
	DefaultName        string
	GatekeeperEndpoint string
	GatekeeperParam    string
	GatekeeperSelector string
	JwtAudience        string
	JwtIssuer          string
	JwtPublicKeyFile   string
}

// GetUserMiddleware returns a middleware that sets user details.
func GetUserMiddleware(authType string, cfg AuthConfig) (mux.MiddlewareFunc, error) {
	// Setup auth; default is all users will be anonymous.
	switch authType {
	case "admin":
		return AdminDefaultMiddleware(cfg.DefaultName), nil
	case "user":
		return UserDefaultMiddleware(cfg.DefaultName), nil
	case "jwt":
		return JWTMiddleware(cfg.JwtAudience, cfg.JwtIssuer, cfg.JwtPublicKeyFile)
	case "kong":
		return KongMiddleware()
	}
	return func(next http.Handler) http.Handler {
		return next
	}, nil
}

// AdminDefaultMiddleware uses a default "admin" context.
func AdminDefaultMiddleware(defaultName string) func(http.Handler) http.Handler {
	return NewUserDefaultMiddleware(func() *User { return NewUser(defaultName).WithRoles("admin") })
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
