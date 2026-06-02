package psql

import (
	"context"
	"database/sql"
	"errors"
	"strings"
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

type userRow struct {
	ID               uuid.UUID      `db:"id"`
	Login            string         `db:"login"`
	PasswordHash     string         `db:"password_hash"`
	FirstName        string         `db:"first_name"`
	LastName         string         `db:"last_name"`
	Patronymic       sql.NullString `db:"patronymic"`
	Email            sql.NullString `db:"email"`
	PhoneNumber      sql.NullString `db:"phone_number"`
	RoleID           uuid.UUID      `db:"role_id"`
	RoleCode         string         `db:"role_code"`
	Status           string         `db:"status"`
	FailedLoginCount int            `db:"failed_login_count"`
	LockedUntil      *time.Time     `db:"locked_until"`
	CreatedAt        time.Time      `db:"created_at"`
	UpdatedAt        time.Time      `db:"updated_at"`
}

func (r *userRow) toModel() *model.User {
	return &model.User{
		ID:               r.ID,
		Login:            vo.NewLoginFromStored(r.Login),
		PasswordHash:     vo.NewPasswordHashFromStored(r.PasswordHash),
		FirstName:        r.FirstName,
		LastName:         r.LastName,
		Patronymic:       r.Patronymic.String,
		Email:            vo.NewEmailFromStored(r.Email.String),
		Phone:            vo.NewPhoneNumberFromStored(r.PhoneNumber.String),
		RoleID:           r.RoleID,
		RoleCode:         r.RoleCode,
		Status:           r.Status,
		FailedLoginCount: r.FailedLoginCount,
		LockedUntil:      r.LockedUntil,
		CreatedAt:        r.CreatedAt,
		UpdatedAt:        r.UpdatedAt,
	}
}

const selectUserByLogin = `
	SELECT u.id, u.login, u.password_hash, u.first_name, u.last_name, u.patronymic,
	       u.email, u.phone_number, u.role_id, r.code AS role_code, u.status,
	       u.failed_login_count, u.locked_until, u.created_at, u.updated_at
	FROM users u
	JOIN roles r ON r.id = u.role_id
	WHERE lower(u.login) = lower($1)`

func (r *userRepository) GetUserByLogin(ctx context.Context, login vo.Login) (*model.User, error) {
	var row userRow
	if err := r.db.GetContext(ctx, &row, selectUserByLogin, login.String()); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, domain.ErrUserNotFound
		}
		return nil, err
	}
	return row.toModel(), nil
}

func (r *userRepository) ExistsByLogin(ctx context.Context, login vo.Login) (bool, error) {
	const query = `SELECT EXISTS(SELECT 1 FROM users WHERE lower(login) = lower($1))`
	var exists bool
	if err := r.db.GetContext(ctx, &exists, query, login.String()); err != nil {
		return false, err
	}
	return exists, nil
}

const insertUser = `
	INSERT INTO users (login, password_hash, first_name, last_name, patronymic, email, phone_number, role_id)
	VALUES ($1, $2, $3, $4, $5, $6, $7, (SELECT id FROM roles WHERE code = 'user'))
	RETURNING id`

func (r *userRepository) CreateUser(ctx context.Context, u *model.User) (uuid.UUID, error) {
	var id uuid.UUID
	err := r.db.QueryRowContext(ctx, insertUser,
		u.Login.String(),
		u.PasswordHash.String(),
		u.FirstName,
		u.LastName,
		nullIfEmpty(u.Patronymic),
		nullIfEmpty(u.Email.String()),
		nullIfEmpty(u.Phone.String()),
	).Scan(&id)
	if err != nil {
		return uuid.Nil, err
	}
	return id, nil
}

func nullIfEmpty(s string) any {
	if strings.TrimSpace(s) == "" {
		return nil
	}
	return s
}
