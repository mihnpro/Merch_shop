package psql

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"

	"github.com/mihnpro/Merch_shop/services/user_customer/internal/domain"
	"github.com/mihnpro/Merch_shop/services/user_customer/internal/domain/model"
	"github.com/mihnpro/Merch_shop/services/user_customer/internal/domain/repository"
	vo "github.com/mihnpro/Merch_shop/services/user_customer/internal/domain/valueobject"
)

type userRepository struct {
	db *sqlx.DB
}

func NewUserRepository(db *sqlx.DB) repository.UserRepository {
	return &userRepository{db: db}
}

func (r *userRepository) GetUserByLogin(ctx context.Context, login vo.Login) (*model.User, error) {
	row, err := getUserByLogin(ctx, r.db, login.String())
	if err != nil {
		return nil, err
	}
	return row.toModel(), nil
}

func (r *userRepository) ExistsByLogin(ctx context.Context, login vo.Login) (bool, error) {
	return existsByLogin(ctx, r.db, login.String())
}

func (r *userRepository) CreateUser(ctx context.Context, u *model.User) (uuid.UUID, error) {
	return insertUser(ctx, r.db, u)
}

func (r *userRepository) GetUserByID(ctx context.Context, id uuid.UUID) (*model.User, error) {
	row, err := getUserByID(ctx, r.db, id)
	if err != nil {
		return nil, err
	}
	return row.toModel(), nil
}

func (r *userRepository) UpdateUserStatus(ctx context.Context, id uuid.UUID, status vo.Status) (*model.User, error) {
	row, err := updateUserStatus(ctx, r.db, id, status.String())
	if err != nil {
		return nil, err
	}
	return row.toModel(), nil
}

func (r *userRepository) UpdateUserRole(ctx context.Context, id uuid.UUID, role vo.Role) (*model.User, error) {
	row, err := updateUserRole(ctx, r.db, id, role.String())
	if err != nil {
		return nil, err
	}
	return row.toModel(), nil
}

func (r *userRepository) UpdateUserProfile(ctx context.Context, id uuid.UUID, p repository.ProfileUpdate) (*model.User, error) {
	row, err := updateUserProfile(ctx, r.db, id, p)
	if err != nil {
		return nil, err
	}
	return row.toModel(), nil
}

func (r *userRepository) CountUsersCreatedSince(ctx context.Context, since time.Time) (int, error) {
	return countUsersCreatedSince(ctx, r.db, since)
}

func (r *userRepository) UpdatePassword(ctx context.Context, id uuid.UUID, hash vo.PasswordHash) error {
	n, err := updatePassword(ctx, r.db, id, hash.String())
	if err != nil {
		return err
	}
	if n == 0 {
		return domain.ErrUserNotFound
	}
	return nil
}

func (r *userRepository) ListUsers(ctx context.Context, f repository.ListUsersFilter) ([]*model.User, string, error) {
	limit := f.PageSize
	if limit <= 0 || limit > 100 {
		limit = 25
	}

	rows, err := selectUsersPage(ctx, r.db, f, limit)
	if err != nil {
		return nil, "", err
	}

	var nextToken string
	if len(rows) > limit {
		rows = rows[:limit]
		last := rows[limit-1]
		nextToken = encodeUserCursor(last.CreatedAt, last.ID)
	}

	var users []*model.User
	for i := range rows {
		users = append(users, rows[i].toModel())
	}
	return users, nextToken, nil
}
