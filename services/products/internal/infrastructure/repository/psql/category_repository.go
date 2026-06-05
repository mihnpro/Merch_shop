package psql

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/lib/pq"

	"github.com/mihnpro/Merch_shop/services/products/internal/domain"
	"github.com/mihnpro/Merch_shop/services/products/internal/domain/model"
	"github.com/mihnpro/Merch_shop/services/products/internal/domain/repository"
	vo "github.com/mihnpro/Merch_shop/services/products/internal/domain/valueobject"
)

const uniqueViolation = "23505"

type categoryRepository struct {
	db *sqlx.DB
}

func NewCategoryRepository(db *sqlx.DB) repository.CategoryRepository {
	return &categoryRepository{db: db}
}

type categoryRow struct {
	ID        uuid.UUID `db:"id"`
	Code      string    `db:"code"`
	Name      string    `db:"name"`
	Active    bool      `db:"active"`
	CreatedAt time.Time `db:"created_at"`
	UpdatedAt time.Time `db:"updated_at"`
}

func (r *categoryRow) toModel() *model.Category {
	return &model.Category{
		ID:        r.ID,
		Code:      vo.NewCategoryCodeFromStored(r.Code),
		Name:      r.Name,
		Active:    r.Active,
		CreatedAt: r.CreatedAt,
		UpdatedAt: r.UpdatedAt,
	}
}

const insertCategory = `
	INSERT INTO categories (code, name)
	VALUES ($1, $2)
	RETURNING id, active, created_at, updated_at`

func (r *categoryRepository) Create(ctx context.Context, c *model.Category) (*model.Category, error) {
	err := r.db.QueryRowContext(ctx, insertCategory, c.Code.String(), c.Name).
		Scan(&c.ID, &c.Active, &c.CreatedAt, &c.UpdatedAt)
	if err != nil {
		var pqErr *pq.Error
		if errors.As(err, &pqErr) && string(pqErr.Code) == uniqueViolation {
			return nil, domain.ErrCategoryAlreadyExists
		}
		return nil, err
	}
	return c, nil
}

const updateCategory = `
	UPDATE categories
	SET name = $1, active = $2, updated_at = NOW()
	WHERE id = $3
	RETURNING updated_at`

func (r *categoryRepository) Update(ctx context.Context, c *model.Category) (*model.Category, error) {
	err := r.db.QueryRowContext(ctx, updateCategory, c.Name, c.Active, c.ID).Scan(&c.UpdatedAt)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, domain.ErrCategoryNotFound
		}
		return nil, err
	}
	return c, nil
}

func (r *categoryRepository) GetByID(ctx context.Context, id uuid.UUID) (*model.Category, error) {
	const q = `SELECT id, code, name, active, created_at, updated_at FROM categories WHERE id = $1`
	var row categoryRow
	if err := r.db.GetContext(ctx, &row, q, id); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, domain.ErrCategoryNotFound
		}
		return nil, err
	}
	return row.toModel(), nil
}

func (r *categoryRepository) List(ctx context.Context, activeOnly bool) ([]*model.Category, error) {
	q := `SELECT id, code, name, active, created_at, updated_at FROM categories`
	if activeOnly {
		q += ` WHERE active`
	}
	q += ` ORDER BY name`

	var rows []categoryRow
	if err := r.db.SelectContext(ctx, &rows, q); err != nil {
		return nil, err
	}
	categories := make([]*model.Category, 0, len(rows))
	for idx := range rows {
		categories = append(categories, rows[idx].toModel())
	}
	return categories, nil
}

func (r *categoryRepository) ExistsByCode(ctx context.Context, code vo.CategoryCode) (bool, error) {
	const q = `SELECT EXISTS(SELECT 1 FROM categories WHERE code = $1)`
	var exists bool
	if err := r.db.GetContext(ctx, &exists, q, code.String()); err != nil {
		return false, err
	}
	return exists, nil
}
