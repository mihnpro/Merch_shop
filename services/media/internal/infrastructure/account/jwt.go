package account

import (
	"fmt"

	"github.com/golang-jwt/jwt/v5"

	"github.com/mihnpro/Merch_shop/services/media/internal/app/dto"
	"github.com/mihnpro/Merch_shop/services/media/internal/app/port"
	"github.com/mihnpro/Merch_shop/services/media/internal/domain"
)

const tokenTypeAccess = "access"

type verifier struct {
	accessSecret []byte
}

func NewVerifier(accessSecret string) port.Verifier {
	return &verifier{accessSecret: []byte(accessSecret)}
}

type claims struct {
	UserID    string `json:"user_id"`
	Role      string `json:"role"`
	TokenType string `json:"token_type"`
	jwt.RegisteredClaims
}

func (v *verifier) ParseAccess(token string) (dto.Identity, error) {
	if token == "" {
		return dto.Identity{}, domain.ErrUnauthenticated
	}

	parsed, err := jwt.ParseWithClaims(token, &claims{}, func(t *jwt.Token) (any, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", t.Header["alg"])
		}
		return v.accessSecret, nil
	}, jwt.WithValidMethods([]string{"HS256"}))
	if err != nil || !parsed.Valid {
		return dto.Identity{}, domain.ErrInvalidToken
	}

	c, ok := parsed.Claims.(*claims)
	if !ok || c.TokenType != tokenTypeAccess {
		return dto.Identity{}, domain.ErrInvalidToken
	}

	return dto.Identity{UserID: c.UserID, Role: c.Role}, nil
}
