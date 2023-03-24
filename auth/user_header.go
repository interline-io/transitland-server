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
				user := newCtxUser(v)
				r = r.WithContext(context.WithValue(r.Context(), userCtxKey, user))
			}
			next.ServeHTTP(w, r)
		})
	}, nil
}
