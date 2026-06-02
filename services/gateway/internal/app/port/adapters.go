package port

import (
	"context"

	"github.com/mihnpro/Merch_shop/services/gateway/internal/app/dto"
)

type UserAdapter interface {
	Register(ctx context.Context, in dto.RegisterInput) (dto.UserView, error)
	Login(ctx context.Context, in dto.LoginInput) (dto.AuthResult, error)
	Refresh(ctx context.Context, refreshToken string) (dto.TokenPair, error)
	Logout(ctx context.Context, refreshToken string) error
}

type Adapters struct {
	User UserAdapter
}
