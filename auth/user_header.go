package auth

import (
	"context"
	"net/http"
)

// UserHeaderMiddleware checks and pulls user ID from specified headers.
func UserHeaderMiddleware(header string) (func(http.Handler) http.Handler, error) {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if v := r.Header.Get(header); v != "" {
				ctx := r.Context()
				user := ForContext(ctx)
				if user == nil {
					user = NewUser(v)
				}
				user.Name = v
				user.AddRoles("user")
				r = r.WithContext(context.WithValue(r.Context(), userCtxKey, user))
			}
			next.ServeHTTP(w, r)
		})
	}, nil
}
