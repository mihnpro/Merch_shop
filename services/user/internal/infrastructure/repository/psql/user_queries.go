package psql

import (
	"context"
	"database/sql"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"

	"github.com/mihnpro/Merch_shop/services/user_customer/internal/domain"
	"github.com/mihnpro/Merch_shop/services/user_customer/internal/domain/model"
	"github.com/mihnpro/Merch_shop/services/user_customer/internal/domain/repository"
	vo "github.com/mihnpro/Merch_shop/services/user_customer/internal/domain/valueobject"
)

type userRow struct {
	ID               uuid.UUID      `db:"id"`
	Login            string         `db:"login"`
	PasswordHash     string         `db:"password_hash"`
	FirstName        string         `db:"first_name"`
	LastName         string         `db:"last_name"`
	Patronymic       sql.NullString `db:"patronymic"`
	Email            sql.NullString `db:"email"`
	PhoneNumber      sql.NullString `db:"phone_number"`
	Role             string         `db:"role"`
	Status           string         `db:"status"`
	FailedLoginCount int            `db:"failed_login_count"`
	LockedUntil      *time.Time     `db:"locked_until"`
	LastLoginAt      *time.Time     `db:"last_login_at"`
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
		Role:             vo.NewRoleFromStored(r.Role),
		Status:           vo.NewStatusFromStored(r.Status),
		FailedLoginCount: r.FailedLoginCount,
		LockedUntil:      r.LockedUntil,
		LastLoginAt:      r.LastLoginAt,
		CreatedAt:        r.CreatedAt,
		UpdatedAt:        r.UpdatedAt,
	}
}

const selectUserCols = `u.id, u.login, u.password_hash, u.first_name, u.last_name, u.patronymic,
	       u.email, u.phone_number, u.role, u.status,
	       u.failed_login_count, u.locked_until, u.last_login_at, u.created_at, u.updated_at`

const (
	selectUserByLoginSQL = `SELECT ` + selectUserCols + ` FROM users u WHERE lower(u.login) = lower($1)`
	selectUserByIDSQL    = `SELECT ` + selectUserCols + ` FROM users u WHERE u.id = $1`
	existsByLoginSQL     = `SELECT EXISTS(SELECT 1 FROM users WHERE lower(login) = lower($1))`
	insertUserSQL        = `INSERT INTO users (login, password_hash, first_name, last_name, patronymic, email, phone_number, role)
	VALUES ($1, $2, $3, $4, $5, $6, $7, 'user')
	RETURNING id`
	updateUserStatusSQL = `UPDATE users SET status = $2 WHERE id = $1 RETURNING ` + selectUserCols
	updateUserRoleSQL   = `UPDATE users SET role = $2 WHERE id = $1 RETURNING ` + selectUserCols
	updatePasswordSQL   = `UPDATE users SET password_hash = $2 WHERE id = $1`
)


func getUserByLogin(ctx context.Context, q sqlx.QueryerContext, login string) (*userRow, error) {
	return getUserRow(ctx, q, selectUserByLoginSQL, login)
}

func getUserByID(ctx context.Context, q sqlx.QueryerContext, id uuid.UUID) (*userRow, error) {
	return getUserRow(ctx, q, selectUserByIDSQL, id)
}

func updateUserStatus(ctx context.Context, q sqlx.QueryerContext, id uuid.UUID, status string) (*userRow, error) {
	return getUserRow(ctx, q, updateUserStatusSQL, id, status)
}

func updateUserRole(ctx context.Context, q sqlx.QueryerContext, id uuid.UUID, role string) (*userRow, error) {
	return getUserRow(ctx, q, updateUserRoleSQL, id, role)
}

func getUserRow(ctx context.Context, q sqlx.QueryerContext, query string, args ...any) (*userRow, error) {
	var row userRow
	if err := sqlx.GetContext(ctx, q, &row, query, args...); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, domain.ErrUserNotFound
		}
		return nil, err
	}
	return &row, nil
}

func existsByLogin(ctx context.Context, q sqlx.QueryerContext, login string) (bool, error) {
	var exists bool
	if err := sqlx.GetContext(ctx, q, &exists, existsByLoginSQL, login); err != nil {
		return false, err
	}
	return exists, nil
}

func insertUser(ctx context.Context, q sqlx.QueryerContext, u *model.User) (uuid.UUID, error) {
	var id uuid.UUID
	err := sqlx.GetContext(ctx, q, &id, insertUserSQL,
		u.Login.String(),
		u.PasswordHash.String(),
		u.FirstName,
		u.LastName,
		nullIfEmpty(u.Patronymic),
		nullIfEmpty(u.Email.String()),
		nullIfEmpty(u.Phone.String()),
	)
	if err != nil {
		return uuid.Nil, err
	}
	return id, nil
}

func updatePassword(ctx context.Context, exec sqlx.ExecerContext, id uuid.UUID, hash string) (int64, error) {
	res, err := exec.ExecContext(ctx, updatePasswordSQL, id, hash)
	if err != nil {
		return 0, err
	}
	n, _ := res.RowsAffected()
	return n, nil
}

func selectUsersPage(ctx context.Context, db *sqlx.DB, f repository.ListUsersFilter, limit int) ([]userRow, error) {
	args := []any{}
	conds := []string{}
	i := 1

	if f.Search != "" {
		conds = append(conds, fmt.Sprintf(
			`(lower(u.login) LIKE lower($%d) OR lower(u.email) LIKE lower($%d) OR lower(u.first_name) LIKE lower($%d) OR lower(u.last_name) LIKE lower($%d))`,
			i, i+1, i+2, i+3,
		))
		pat := "%" + f.Search + "%"
		args = append(args, pat, pat, pat, pat)
		i += 4
	}
	if f.RoleCode != "" {
		conds = append(conds, fmt.Sprintf(`u.role = $%d`, i))
		args = append(args, f.RoleCode)
		i++
	}
	if f.Status != "" {
		conds = append(conds, fmt.Sprintf(`u.status = $%d`, i))
		args = append(args, f.Status)
		i++
	}
	if f.PageToken != "" {
		cur, err := decodeUserCursor(f.PageToken)
		if err == nil {
			conds = append(conds, fmt.Sprintf(`(u.created_at, u.id) < ($%d, $%d)`, i, i+1))
			args = append(args, cur.CreatedAt, cur.ID)
			i += 2
		}
	}

	where := ""
	if len(conds) > 0 {
		where = "WHERE " + strings.Join(conds, " AND ")
	}

	query := fmt.Sprintf(`SELECT `+selectUserCols+` FROM users u %s ORDER BY u.created_at DESC, u.id DESC LIMIT $%d`, where, i)
	args = append(args, limit+1)

	var rows []userRow
	if err := db.SelectContext(ctx, &rows, query, args...); err != nil {
		return nil, err
	}
	return rows, nil
}

func nullIfEmpty(s string) any {
	if strings.TrimSpace(s) == "" {
		return nil
	}
	return s
}


type userCursor struct {
	CreatedAt time.Time `json:"ca"`
	ID        uuid.UUID `json:"id"`
}

func encodeUserCursor(createdAt time.Time, id uuid.UUID) string {
	b, _ := json.Marshal(userCursor{CreatedAt: createdAt, ID: id})
	return base64.StdEncoding.EncodeToString(b)
}

func decodeUserCursor(token string) (userCursor, error) {
	b, err := base64.StdEncoding.DecodeString(token)
	if err != nil {
		return userCursor{}, err
	}
	var cur userCursor
	if err := json.Unmarshal(b, &cur); err != nil {
		return userCursor{}, err
	}
	return cur, nil
}
