package auth

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestKongMiddleware(t *testing.T) {
	tcs := []struct {
		name       string
		consumerId string
		code       int
		user       User
	}{
		{"test", "test@transitland", 200, newCtxUser("test@transitland").WithRoles("user")},
		{"no user", "", 200, nil},
	}
	for _, tc := range tcs {
		t.Run(tc.name, func(t *testing.T) {
			mf, err := UserHeaderMiddleware("x-consumer-username")
			if err != nil {
				t.Fatal(err)
			}
			req := httptest.NewRequest(http.MethodGet, "/", nil)
			if tc.consumerId != "" {
				req.Header.Add("x-consumer-username", tc.consumerId)
			}
			testAuthMiddleware(t, req, mf, tc.code, tc.user)
		})
	}
}
