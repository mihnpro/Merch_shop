package psql

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/lib/pq"

	"github.com/mihnpro/Merch_shop/services/products/internal/domain"
	"github.com/mihnpro/Merch_shop/services/products/internal/domain/model"
	"github.com/mihnpro/Merch_shop/services/products/internal/domain/repository"
	vo "github.com/mihnpro/Merch_shop/services/products/internal/domain/valueobject"
)

type productRepository struct {
	db *sqlx.DB
}

func NewProductRepository(db *sqlx.DB) repository.ProductRepository {
	return &productRepository{db: db}
}

type productRow struct {
	ID          uuid.UUID      `db:"id"`
	Name        string         `db:"name"`
	Description string         `db:"description"`
	PricePoints int64          `db:"price_points"`
	Sizes       pq.StringArray `db:"sizes"`
	PhotoKeys   pq.StringArray `db:"photo_keys"`
	Active      bool           `db:"active"`
	Version     int            `db:"version"`
	CreatedAt   time.Time      `db:"created_at"`
	UpdatedAt   time.Time      `db:"updated_at"`

	CategoryID        uuid.UUID `db:"category_id"`
	CategoryCode      string    `db:"category_code"`
	CategoryName      string    `db:"category_name"`
	CategoryActive    bool      `db:"category_active"`
	CategoryCreatedAt time.Time `db:"category_created_at"`
	CategoryUpdatedAt time.Time `db:"category_updated_at"`
}

func (r *productRow) toModel() *model.Product {
	sizes := make([]vo.SizeCode, 0, len(r.Sizes))
	for _, s := range r.Sizes {
		sizes = append(sizes, vo.NewSizeCodeFromStored(s))
	}
	photos := make([]vo.PhotoKey, 0, len(r.PhotoKeys))
	for _, k := range r.PhotoKeys {
		photos = append(photos, vo.NewPhotoKeyFromStored(k))
	}
	return &model.Product{
		ID:          r.ID,
		Name:        r.Name,
		Description: r.Description,
		Price:       vo.NewPricePointsFromStored(r.PricePoints),
		Category: model.Category{
			ID:        r.CategoryID,
			Code:      vo.NewCategoryCodeFromStored(r.CategoryCode),
			Name:      r.CategoryName,
			Active:    r.CategoryActive,
			CreatedAt: r.CategoryCreatedAt,
			UpdatedAt: r.CategoryUpdatedAt,
		},
		Sizes:     sizes,
		PhotoKeys: photos,
		Active:    r.Active,
		Version:   r.Version,
		CreatedAt: r.CreatedAt,
		UpdatedAt: r.UpdatedAt,
	}
}

const selectProductColumns = `
	p.id, p.name, p.description, p.price_points, p.sizes, p.photo_keys,
	p.active, p.version, p.created_at, p.updated_at,
	c.id AS category_id, c.code AS category_code, c.name AS category_name,
	c.active AS category_active, c.created_at AS category_created_at, c.updated_at AS category_updated_at`

const insertProduct = `
	INSERT INTO products (name, description, price_points, category_id, sizes, photo_keys)
	VALUES ($1, $2, $3, $4, $5, $6)
	RETURNING id, active, version, created_at, updated_at`

func (r *productRepository) Create(ctx context.Context, p *model.Product) (*model.Product, error) {
	err := r.db.QueryRowContext(ctx, insertProduct,
		p.Name,
		p.Description,
		p.Price.Int64(),
		p.Category.ID,
		pq.Array(p.SizeCodes()),
		pq.Array(p.PhotoKeyStrings()),
	).Scan(&p.ID, &p.Active, &p.Version, &p.CreatedAt, &p.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return p, nil
}

func (r *productRepository) GetByID(ctx context.Context, id uuid.UUID) (*model.Product, error) {
	query := fmt.Sprintf(`
		SELECT %s
		FROM products p
		JOIN categories c ON c.id = p.category_id
		WHERE p.id = $1`, selectProductColumns)

	var row productRow
	if err := r.db.GetContext(ctx, &row, query, id); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, domain.ErrProductNotFound
		}
		return nil, err
	}
	return row.toModel(), nil
}

const updateProduct = `
	UPDATE products
	SET name = $1, description = $2, price_points = $3, category_id = $4,
	    sizes = $5, photo_keys = $6, active = $7, version = version + 1, updated_at = NOW()
	WHERE id = $8 AND version = $9
	RETURNING version, updated_at`

func (r *productRepository) Update(ctx context.Context, p *model.Product, expectedVersion int) (*model.Product, error) {
	err := r.db.QueryRowContext(ctx, updateProduct,
		p.Name,
		p.Description,
		p.Price.Int64(),
		p.Category.ID,
		pq.Array(p.SizeCodes()),
		pq.Array(p.PhotoKeyStrings()),
		p.Active,
		p.ID,
		expectedVersion,
	).Scan(&p.Version, &p.UpdatedAt)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, r.conflictOrNotFound(ctx, p.ID)
		}
		return nil, err
	}
	return p, nil
}

func (r *productRepository) conflictOrNotFound(ctx context.Context, id uuid.UUID) error {
	const q = `SELECT EXISTS(SELECT 1 FROM products WHERE id = $1)`
	var exists bool
	if err := r.db.GetContext(ctx, &exists, q, id); err != nil {
		return err
	}
	if !exists {
		return domain.ErrProductNotFound
	}
	return domain.ErrVersionConflict
}

func (r *productRepository) Deactivate(ctx context.Context, id uuid.UUID) error {
	const q = `UPDATE products SET active = false, updated_at = NOW() WHERE id = $1`
	res, err := r.db.ExecContext(ctx, q, id)
	if err != nil {
		return err
	}
	affected, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if affected == 0 {
		return domain.ErrProductNotFound
	}
	return nil
}

func (r *productRepository) List(ctx context.Context, f repository.ProductFilter) (repository.ProductPage, error) {
	conds := []string{}
	args := []any{}
	i := 1

	if f.ActiveOnly {
		conds = append(conds, "p.active")
	}
	if f.CategoryID != nil {
		conds = append(conds, fmt.Sprintf("p.category_id = $%d", i))
		args = append(args, *f.CategoryID)
		i++
	}
	if strings.TrimSpace(f.Search) != "" {
		conds = append(conds, fmt.Sprintf("p.name ILIKE $%d", i))
		args = append(args, "%"+strings.TrimSpace(f.Search)+"%")
		i++
	}
	if f.Cursor != nil {
		conds = append(conds, fmt.Sprintf("(p.created_at, p.id) < ($%d, $%d)", i, i+1))
		args = append(args, f.Cursor.CreatedAt, f.Cursor.ID)
		i += 2
	}

	where := ""
	if len(conds) > 0 {
		where = "WHERE " + strings.Join(conds, " AND ")
	}

	limit := f.Limit
	if limit <= 0 {
		limit = 24
	}

	query := fmt.Sprintf(`
		SELECT %s
		FROM products p
		JOIN categories c ON c.id = p.category_id
		%s
		ORDER BY p.created_at DESC, p.id DESC
		LIMIT $%d`, selectProductColumns, where, i)
	args = append(args, limit+1)

	var rows []productRow
	if err := r.db.SelectContext(ctx, &rows, query, args...); err != nil {
		return repository.ProductPage{}, err
	}

	var next *repository.ProductCursor
	if len(rows) > limit {
		last := rows[limit-1]
		next = &repository.ProductCursor{CreatedAt: last.CreatedAt, ID: last.ID}
		rows = rows[:limit]
	}

	products := make([]*model.Product, 0, len(rows))
	for idx := range rows {
		products = append(products, rows[idx].toModel())
	}
	return repository.ProductPage{Products: products, Next: next}, nil
}
