package auth

import (
	"context"
	"net/http"

	"github.com/gorilla/mux"
)

type AuthConfig struct {
	JwtAudience      string
	JwtIssuer        string
	JwtPublicKeyFile string
}

// GetUserMiddleware returns a middleware that sets user details.
func GetUserMiddleware(authType string, cfg AuthConfig) (mux.MiddlewareFunc, error) {
	// Setup auth; default is all users will be anonymous.
	if authType == "admin" {
		return AdminDefaultMiddleware(), nil
	} else if authType == "user" {
		return UserDefaultMiddleware(), nil
	} else if authType == "jwt" {
		return JWTMiddleware(cfg.JwtAudience, cfg.JwtIssuer, cfg.JwtPublicKeyFile)
	}
	return func(next http.Handler) http.Handler {
		return next
	}, nil
}

// AdminDefaultMiddleware uses a default "admin" context.
func AdminDefaultMiddleware() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			user := &User{
				Name:    "",
				IsAnon:  false,
				IsUser:  true,
				IsAdmin: true,
			}
			ctx := context.WithValue(r.Context(), userCtxKey, user)
			r = r.WithContext(ctx)
			next.ServeHTTP(w, r)
		})
	}
}

// UserDefaultMiddleware uses a default "user" context.
func UserDefaultMiddleware() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			user := &User{
				Name:    "",
				IsAnon:  false,
				IsUser:  true,
				IsAdmin: false,
			}
			ctx := context.WithValue(r.Context(), userCtxKey, user)
			r = r.WithContext(ctx)
			next.ServeHTTP(w, r)
		})
	}
}

// AdminRequired limits a request to admin privileges.
func AdminRequired(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		user := ForContext(ctx)
		if user == nil || !user.IsAdmin {
			http.Error(w, "permission denied", http.StatusUnauthorized)
			return
		}
		next.ServeHTTP(w, r)
	})
}

// UserRequired limits a request to user privileges.
func UserRequired(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		user := ForContext(ctx)
		if user == nil || !user.IsUser {
			http.Error(w, "permission denied", http.StatusUnauthorized)
			return
		}
		next.ServeHTTP(w, r)
	})
}
