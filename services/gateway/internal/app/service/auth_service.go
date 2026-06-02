package service

import (
	"context"

	"github.com/mihnpro/Merch_shop/services/gateway/internal/app/dto"
	"github.com/mihnpro/Merch_shop/services/gateway/internal/app/port"
)

type AuthQueryService interface {
	Register(ctx context.Context, in dto.RegisterInput) (dto.UserView, error)
	Login(ctx context.Context, in dto.LoginInput) (dto.AuthResult, error)
	Refresh(ctx context.Context, refreshToken string) (dto.TokenPair, error)
	Logout(ctx context.Context, refreshToken string) error
}

type authQueryService struct {
	users port.UserAdapter
}

func NewAuthQueryService(users port.UserAdapter) AuthQueryService {
	return &authQueryService{users: users}
}

func (s *authQueryService) Register(ctx context.Context, in dto.RegisterInput) (dto.UserView, error) {
	return s.users.Register(ctx, in)
}

func (s *authQueryService) Login(ctx context.Context, in dto.LoginInput) (dto.AuthResult, error) {
	return s.users.Login(ctx, in)
}

func (s *authQueryService) Refresh(ctx context.Context, refreshToken string) (dto.TokenPair, error) {
	return s.users.Refresh(ctx, refreshToken)
}

func (s *authQueryService) Logout(ctx context.Context, refreshToken string) error {
	return s.users.Logout(ctx, refreshToken)
}
