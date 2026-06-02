package account

import (
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"

	"github.com/mihnpro/Merch_shop/services/user_customer/internal/app/port"
	"github.com/mihnpro/Merch_shop/services/user_customer/internal/domain"
	"github.com/mihnpro/Merch_shop/services/user_customer/internal/domain/model"
)

const (
	tokenTypeAccess  = "access"
	tokenTypeRefresh = "refresh"
	tokenIssuer      = "UserService"
)

type claims struct {
	UserID    uuid.UUID `json:"user_id"`
	Email     string    `json:"email"`
	Role      string    `json:"role"`
	TokenType string    `json:"token_type"`
	jwt.RegisteredClaims
}

func (a *account) GenerateTokens(identity model.Identity) (port.TokenPair, error) {
	now := time.Now()
	accessExp := now.Add(a.accessTTL)
	refreshExp := now.Add(a.refreshTTL)

	accessToken, err := a.signToken(identity, tokenTypeAccess, now, accessExp, a.accessSecret)
	if err != nil {
		return port.TokenPair{}, err
	}
	refreshToken, err := a.signToken(identity, tokenTypeRefresh, now, refreshExp, a.refreshSecret)
	if err != nil {
		return port.TokenPair{}, err
	}

	return port.TokenPair{
		AccessToken:      accessToken,
		RefreshToken:     refreshToken,
		AccessExpiresAt:  accessExp,
		RefreshExpiresAt: refreshExp,
	}, nil
}

func (a *account) signToken(identity model.Identity, tokenType string, issuedAt, expiresAt time.Time, secret string) (string, error) {
	c := &claims{
		UserID:    identity.UserID,
		Email:     identity.Email,
		Role:      identity.Role,
		TokenType: tokenType,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expiresAt),
			IssuedAt:  jwt.NewNumericDate(issuedAt),
			Issuer:    tokenIssuer,
			Subject:   identity.UserID.String(),
		},
	}
	return jwt.NewWithClaims(jwt.SigningMethodHS256, c).SignedString([]byte(secret))
}

func (a *account) ValidateToken(tokenString string) (model.Identity, error) {

	parsed, _, err := jwt.NewParser().ParseUnverified(tokenString, &claims{})
	if err != nil {
		return model.Identity{}, domain.ErrInvalidToken
	}
	unverified, ok := parsed.Claims.(*claims)
	if !ok {
		return model.Identity{}, domain.ErrInvalidToken
	}

	var secret string
	switch unverified.TokenType {
	case tokenTypeAccess:
		secret = a.accessSecret
	case tokenTypeRefresh:
		secret = a.refreshSecret
	default:
		return model.Identity{}, domain.ErrInvalidToken
	}

	verified, err := jwt.ParseWithClaims(tokenString, &claims{}, func(t *jwt.Token) (any, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, domain.ErrInvalidToken
		}
		return []byte(secret), nil
	}, jwt.WithValidMethods([]string{"HS256"}))
	if err != nil || !verified.Valid {
		return model.Identity{}, domain.ErrInvalidToken
	}

	validClaims, ok := verified.Claims.(*claims)
	if !ok {
		return model.Identity{}, domain.ErrInvalidToken
	}

	return model.Identity{
		UserID: validClaims.UserID,
		Email:  validClaims.Email,
		Role:   validClaims.Role,
	}, nil
}
