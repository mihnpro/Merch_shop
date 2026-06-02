package account

import (
	"errors"
	"fmt"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"

	"github.com/mihnpro/Merch_shop/services/gateway/internal/app/port"
	"github.com/mihnpro/Merch_shop/services/gateway/internal/domain"
)

type verifier struct {
	accessSecret  []byte
	refreshSecret []byte
}

func NewVerifier(accessSecret, refreshSecret string) port.Verifier {
	return &verifier{
		accessSecret:  []byte(accessSecret),
		refreshSecret: []byte(refreshSecret),
	}
}

type rawClaims struct {
	UserID    string `json:"user_id"`
	Role      string `json:"role"`
	TokenType string `json:"token_type"`
	jwt.RegisteredClaims
}

func (v *verifier) VerifyAccess(token string) (*domain.Claims, error) {
	return v.verify(token, v.accessSecret, "access")
}

func (v *verifier) VerifyRefresh(token string) (*domain.Claims, error) {
	return v.verify(token, v.refreshSecret, "refresh")
}

func (v *verifier) verify(tokenStr string, secret []byte, expectedType string) (*domain.Claims, error) {
	tok, err := jwt.ParseWithClaims(tokenStr, &rawClaims{}, func(t *jwt.Token) (any, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", t.Header["alg"])
		}
		return secret, nil
	}, jwt.WithValidMethods([]string{"HS256"}))
	if err != nil || !tok.Valid {
		return nil, errors.New("invalid token")
	}

	rc, ok := tok.Claims.(*rawClaims)
	if !ok {
		return nil, errors.New("invalid claims")
	}
	if rc.TokenType != expectedType {
		return nil, fmt.Errorf("expected %s token, got %s", expectedType, rc.TokenType)
	}

	uid, err := uuid.Parse(rc.UserID)
	if err != nil {
		return nil, errors.New("invalid user_id in token")
	}

	return &domain.Claims{
		UserID:    uid,
		Role:      domain.Role(rc.Role),
		TokenType: rc.TokenType,
	}, nil
}
