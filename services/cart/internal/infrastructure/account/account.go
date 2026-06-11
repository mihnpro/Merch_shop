package account

import (
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"

	"github.com/mihnpro/Merch_shop/services/cart/internal/domain"
	"github.com/mihnpro/Merch_shop/services/cart/internal/domain/model"
)

const tokenTypeAccess = "access"

type claims struct {
	UserID    uuid.UUID `json:"user_id"`
	Email     string    `json:"email"`
	Role      string    `json:"role"`
	TokenType string    `json:"token_type"`
	jwt.RegisteredClaims
}

type Validator struct {
	accessSecret string
}

func NewValidator(accessSecret string) *Validator {
	return &Validator{accessSecret: accessSecret}
}

func (v *Validator) ValidateToken(tokenString string) (model.Identity, error) {
	verified, err := jwt.ParseWithClaims(tokenString, &claims{}, func(t *jwt.Token) (any, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, domain.ErrInvalidToken
		}
		return []byte(v.accessSecret), nil
	}, jwt.WithValidMethods([]string{"HS256"}))
	if err != nil || !verified.Valid {
		return model.Identity{}, domain.ErrInvalidToken
	}

	c, ok := verified.Claims.(*claims)
	if !ok || c.TokenType != tokenTypeAccess {
		return model.Identity{}, domain.ErrInvalidToken
	}

	return model.Identity{
		UserID: c.UserID,
		Email:  c.Email,
		Role:   c.Role,
	}, nil
}
