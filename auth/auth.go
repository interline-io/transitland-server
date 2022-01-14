package auth

import (
	"context"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/interline-io/transitland-server/config"
	"github.com/jmoiron/sqlx"
)

// AdminAuthMiddleware stores the user context, but always as admin
func AdminAuthMiddleware(db sqlx.Ext) (func(http.Handler) http.Handler, error) {
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
	}, nil
}

// UserAuthMiddleware stores a user context.
func UserAuthMiddleware(db sqlx.Ext) (func(http.Handler) http.Handler, error) {
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
	}, nil
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

// GetUserMiddleware returns an authentication middleware for the given configuration.
func GetUserMiddleware(cfg config.Config) mux.MiddlewareFunc {
	// Setup auth; default is all users will be anonymous.
	if cfg.UseAuth == "admin" {
		a, err := AdminAuthMiddleware(nil)
		if err != nil {
			panic(err)
		}
		return a
	} else if cfg.UseAuth == "user" {
		a, err := UserAuthMiddleware(nil)
		if err != nil {
			panic(err)
		}
		return a
	} else if cfg.UseAuth == "jwt" {
		a, err := JWTMiddleware(cfg)
		if err != nil {
			panic(err)
		}
		return a
	}
	return func(next http.Handler) http.Handler {
		return next
	}
}
