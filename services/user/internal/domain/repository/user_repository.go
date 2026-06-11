package repository

import (
	"context"

	"github.com/google/uuid"

	"github.com/mihnpro/Merch_shop/services/user_customer/internal/domain/model"
	vo "github.com/mihnpro/Merch_shop/services/user_customer/internal/domain/valueobject"
)

type ListUsersFilter struct {
	Search    string
	RoleCode  string
	Status    string
	PageSize  int
	PageToken string
}

type UserRepository interface {
	CreateUser(ctx context.Context, u *model.User) (uuid.UUID, error)
	GetUserByLogin(ctx context.Context, login vo.Login) (*model.User, error)
	GetUserByID(ctx context.Context, id uuid.UUID) (*model.User, error)
	ExistsByLogin(ctx context.Context, login vo.Login) (bool, error)
	UpdateUserStatus(ctx context.Context, id uuid.UUID, status vo.Status) (*model.User, error)
	UpdateUserRole(ctx context.Context, id uuid.UUID, role vo.Role) (*model.User, error)
	UpdatePassword(ctx context.Context, id uuid.UUID, hash vo.PasswordHash) error
	ListUsers(ctx context.Context, f ListUsersFilter) ([]*model.User, string, error)
}
