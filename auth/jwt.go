package auth

import (
	"context"
	"crypto/rsa"
	"errors"
	"io/ioutil"
	"net/http"
	"strings"

	"github.com/form3tech-oss/jwt-go"
	"github.com/interline-io/transitland-lib/log"
)

// JWTMiddleware checks and pulls user information from JWT in Authorization header.
func JWTMiddleware(jwtAudience string, jwtIssuer string, pubKeyPath string) (func(http.Handler) http.Handler, error) {
	var verifyKey *rsa.PublicKey
	verifyBytes, err := ioutil.ReadFile(pubKeyPath)
	if err != nil {
		return nil, err
	}
	verifyKey, err = jwt.ParseRSAPublicKeyFromPEM(verifyBytes)
	if err != nil {
		return nil, err
	}
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if tokenString := strings.Split(r.Header.Get("Authorization"), "Bearer "); len(tokenString) == 2 {
				jwtUser, err := validateJwt(verifyKey, jwtAudience, jwtIssuer, tokenString[1])
				if err != nil {
					log.Error().Err(err).Msgf("invalid jwt token")
					http.Error(w, "Unauthorized", http.StatusUnauthorized)
					return
				}
				ctx := r.Context()
				user := ForContext(ctx)
				if user == nil {
					user = NewUser(jwtUser.Name)
				}
				user.Merge(jwtUser)
				r = r.WithContext(context.WithValue(r.Context(), userCtxKey, user))
			}
			next.ServeHTTP(w, r)
		})
	}, nil
}

type CustomClaimsExample struct {
	jwt.StandardClaims
}

func (c *CustomClaimsExample) Valid() error {
	return nil
}

func validateJwt(rsaPublicKey *rsa.PublicKey, jwtAudience string, jwtIssuer string, tokenString string) (*User, error) {
	// Parse the token
	token, err := jwt.ParseWithClaims(tokenString, &CustomClaimsExample{}, func(token *jwt.Token) (interface{}, error) {
		return rsaPublicKey, nil
	})
	if err != nil {
		return nil, err
	}
	claims := token.Claims.(*CustomClaimsExample)
	if !claims.VerifyAudience(jwtAudience, true) {
		return nil, errors.New("invalid audience")
	}
	if !claims.VerifyIssuer(jwtIssuer, true) {
		return nil, errors.New("invalid issuer")
	}
	user := NewUser(claims.Subject).WithRoles("user")
	return user, nil
}
