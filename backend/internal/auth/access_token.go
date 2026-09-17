// Package auth issues and verifies the two tokens a session is made of:
// a short-lived JWT access token (self-contained, verified without a DB
// round-trip) and an opaque, long-lived refresh token (random, stored
// hashed in the database — see refresh_token.go — so a leaked database
// row alone can't be replayed as a token). AuthService (internal/service)
// owns *issuing* both as a pair and rotating them; this package only
// knows how to mint and verify the access token, and how to generate and
// hash the refresh token's random bytes.
package auth

import (
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// AccessClaims is what an access token asserts about its holder.
type AccessClaims struct {
	UserID      string
	IsAnonymous bool
}

// AccessTokenIssuer mints and verifies HS256 access tokens.
type AccessTokenIssuer struct {
	secret []byte
	ttl    time.Duration
}

func NewAccessTokenIssuer(secret string, ttl time.Duration) *AccessTokenIssuer {
	return &AccessTokenIssuer{secret: []byte(secret), ttl: ttl}
}

type accessTokenClaims struct {
	IsAnonymous bool `json:"is_anonymous"`
	jwt.RegisteredClaims
}

func (i *AccessTokenIssuer) Issue(userID string, isAnonymous bool) (string, error) {
	claims := accessTokenClaims{
		IsAnonymous: isAnonymous,
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   userID,
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(i.ttl)),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(i.secret)
}

// Verify returns the claims of a valid, unexpired access token.
func (i *AccessTokenIssuer) Verify(tokenString string) (AccessClaims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &accessTokenClaims{}, func(t *jwt.Token) (any, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", t.Header["alg"])
		}
		return i.secret, nil
	})
	if err != nil {
		return AccessClaims{}, err
	}

	claims, ok := token.Claims.(*accessTokenClaims)
	if !ok || !token.Valid || claims.Subject == "" {
		return AccessClaims{}, fmt.Errorf("invalid token")
	}
	return AccessClaims{UserID: claims.Subject, IsAnonymous: claims.IsAnonymous}, nil
}
