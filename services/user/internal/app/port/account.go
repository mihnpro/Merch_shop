package port

import (
	"time"

	"github.com/mihnpro/Merch_shop/services/user_customer/internal/domain/model"
	vo "github.com/mihnpro/Merch_shop/services/user_customer/internal/domain/valueobject"
)

type TokenPair struct {
	AccessToken      string
	RefreshToken     string
	AccessExpiresAt  time.Time
	RefreshExpiresAt time.Time
}

type Account interface {
	HashPassword(password vo.PlainPassword) (vo.PasswordHash, error)
	VerifyPassword(password vo.PlainPassword, hash vo.PasswordHash) bool
	GenerateTokens(identity model.Identity) (TokenPair, error)
	ValidateToken(token string) (model.Identity, error)
}
