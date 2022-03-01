package jwt

import (
	"gopkg.in/square/go-jose.v2/jwt"
)

type Claims struct {
	jwt.Claims
	Groups []string `json:"groups"`
	Scopes []string `json:"scp"`
}

func TokenClaims(rawToken string) (*Claims, error) {
	token, err := jwt.ParseSigned(rawToken)
	if err != nil {
		return nil, err
	}

	c := &Claims{}
	if err := token.UnsafeClaimsWithoutVerification(c); err != nil {
		return nil, err
	}

	return c, nil
}
