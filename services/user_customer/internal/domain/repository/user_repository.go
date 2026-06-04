package repository

import (
	"context"

	"github.com/google/uuid"

	"github.com/mihnpro/Merch_shop/services/user_customer/internal/domain/model"
	vo "github.com/mihnpro/Merch_shop/services/user_customer/internal/domain/valueobject"
)

type UserRepository interface {
	CreateUser(ctx context.Context, u *model.User) (uuid.UUID, error)
	GetUserByLogin(ctx context.Context, login vo.Login) (*model.User, error)
	ExistsByLogin(ctx context.Context, login vo.Login) (bool, error)
}
