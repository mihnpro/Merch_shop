package read

import (
	"context"

	"github.com/mihnpro/Merch_shop/services/user_customer/internal/domain/model"
	"github.com/mihnpro/Merch_shop/services/user_customer/internal/domain/repository"
	vo "github.com/mihnpro/Merch_shop/services/user_customer/internal/domain/valueobject"
)

type AuthReadService interface {
	GetUserByLogin(ctx context.Context, login string) (*model.User, error)
}

type authReadService struct {
	users repository.UserRepository
}

func NewAuthReadService(users repository.UserRepository) AuthReadService {
	return &authReadService{users: users}
}

func (s *authReadService) GetUserByLogin(ctx context.Context, login string) (*model.User, error) {
	loginVO, err := vo.NewLogin(login)
	if err != nil {
		return nil, err
	}
	return s.users.GetUserByLogin(ctx, loginVO)
}
