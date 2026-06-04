package account

import (
	"errors"
	"fmt"

	"github.com/golang-jwt/jwt/v5"

	"github.com/mihnpro/Merch_shop/services/gateway/internal/app/port"
)

var errInvalidToken = errors.New("invalid token")

type verifier struct {
	accessSecret []byte
}

func NewVerifier(accessSecret string) port.Verifier {
	return &verifier{accessSecret: []byte(accessSecret)}
}

type rawClaims struct {
	TokenType string `json:"token_type"`
	jwt.RegisteredClaims
}

func (v *verifier) VerifyAccess(token string) error {
	tok, err := jwt.ParseWithClaims(token, &rawClaims{}, func(t *jwt.Token) (any, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", t.Header["alg"])
		}
		return v.accessSecret, nil
	}, jwt.WithValidMethods([]string{"HS256"}))
	if err != nil || !tok.Valid {
		return errInvalidToken
	}

	rc, ok := tok.Claims.(*rawClaims)
	if !ok || rc.TokenType != "access" {
		return errInvalidToken
	}
	return nil
}
