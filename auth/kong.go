package auth

import (
	"context"
	"net/http"
)

// KongMiddleware checks and pulls user information from Kong X-Consumer-* headers.
func KongMiddleware() (func(http.Handler) http.Handler, error) {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			var user *User
			if v := r.Header.Get("x-consumer-username"); v != "" {
				user = &User{
					Name:    v,
					IsAnon:  false,
					IsUser:  true,
					IsAdmin: false,
				}
			}
			ctx := context.WithValue(r.Context(), userCtxKey, user)
			r = r.WithContext(ctx)
			next.ServeHTTP(w, r)
		})
	}, nil
}
