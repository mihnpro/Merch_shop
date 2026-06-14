package psql

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"

	"github.com/mihnpro/Merch_shop/services/invetory/internal/domain"
	"github.com/mihnpro/Merch_shop/services/invetory/internal/domain/model"
	"github.com/mihnpro/Merch_shop/services/invetory/internal/domain/repository"
)

type stockRepository struct {
	db *sqlx.DB
}

func NewStockRepository(db *sqlx.DB) repository.StockRepository {
	return &stockRepository{db: db}
}

type stockRow struct {
	ProductID uuid.UUID `db:"product_id"`
	Available int       `db:"available"`
	Version   int       `db:"version"`
	CreatedAt time.Time `db:"created_at"`
	UpdatedAt time.Time `db:"updated_at"`
}

func (r *stockRow) toModel() *model.Stock {
	return &model.Stock{
		ProductID: r.ProductID,
		Available: r.Available,
		Version:   r.Version,
		CreatedAt: r.CreatedAt,
		UpdatedAt: r.UpdatedAt,
	}
}

const selectStock = `
	SELECT product_id, available, version, created_at, updated_at
	FROM stock
	WHERE product_id = $1`

const listStock = `
	SELECT product_id, available, version, created_at, updated_at
	FROM stock
	ORDER BY product_id`

func (r *stockRepository) List(ctx context.Context) ([]*model.Stock, error) {
	var rows []stockRow
	if err := r.db.SelectContext(ctx, &rows, listStock); err != nil {
		return nil, err
	}
	stocks := make([]*model.Stock, 0, len(rows))
	for i := range rows {
		stocks = append(stocks, rows[i].toModel())
	}
	return stocks, nil
}

func (r *stockRepository) GetByProductID(ctx context.Context, productID uuid.UUID) (*model.Stock, error) {
	var row stockRow
	if err := r.db.GetContext(ctx, &row, selectStock, productID); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, domain.ErrStockNotFound
		}
		return nil, err
	}
	return row.toModel(), nil
}

func (r *stockRepository) GetByProductIDs(ctx context.Context, productIDs []uuid.UUID) ([]*model.Stock, error) {
	if len(productIDs) == 0 {
		return []*model.Stock{}, nil
	}
	query, args, err := sqlx.In(`
		SELECT product_id, available, version, created_at, updated_at
		FROM stock
		WHERE product_id IN (?)`, productIDs)
	if err != nil {
		return nil, err
	}
	query = r.db.Rebind(query)

	var rows []stockRow
	if err := r.db.SelectContext(ctx, &rows, query, args...); err != nil {
		return nil, err
	}
	stocks := make([]*model.Stock, 0, len(rows))
	for i := range rows {
		stocks = append(stocks, rows[i].toModel())
	}
	return stocks, nil
}
const insertAdjustment = `
	INSERT INTO stock_adjustments (operation_id, product_id, delta, reason)
	VALUES ($1, $2, $3, $4)
	ON CONFLICT (operation_id) DO NOTHING
	RETURNING id`

const ensureStock = `
	INSERT INTO stock (product_id, available)
	VALUES ($1, 0)
	ON CONFLICT (product_id) DO NOTHING`


const applyDelta = `
	UPDATE stock
	SET available = available + $2, version = version + 1, updated_at = NOW()
	WHERE product_id = $1 AND available + $2 >= 0
	RETURNING product_id, available, version, created_at, updated_at`

func (r *stockRepository) Adjust(ctx context.Context, adj *model.StockAdjustment) (*model.Stock, error) {
	tx, err := r.db.BeginTxx(ctx, nil)
	if err != nil {
		return nil, err
	}
	defer func() { _ = tx.Rollback() }()

	var adjustmentID uuid.UUID
	err = tx.QueryRowContext(ctx, insertAdjustment,
		adj.OperationID, adj.ProductID, adj.Delta.Int(), adj.Reason.String(),
	).Scan(&adjustmentID)

	if errors.Is(err, sql.ErrNoRows) {
		stock, gerr := getStockTx(ctx, tx, adj.ProductID)
		if gerr != nil {
			return nil, gerr
		}
		if cerr := tx.Commit(); cerr != nil {
			return nil, cerr
		}
		return stock, nil
	}
	if err != nil {
		return nil, err
	}

	if _, err := tx.ExecContext(ctx, ensureStock, adj.ProductID); err != nil {
		return nil, err
	}

	var row stockRow
	err = tx.QueryRowContext(ctx, applyDelta, adj.ProductID, adj.Delta.Int()).
		Scan(&row.ProductID, &row.Available, &row.Version, &row.CreatedAt, &row.UpdatedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, domain.ErrInsufficientStock
	}
	if err != nil {
		return nil, err
	}

	if err := tx.Commit(); err != nil {
		return nil, err
	}
	return row.toModel(), nil
}

func getStockTx(ctx context.Context, tx *sqlx.Tx, productID uuid.UUID) (*model.Stock, error) {
	var row stockRow
	if err := tx.GetContext(ctx, &row, selectStock, productID); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, domain.ErrStockNotFound
		}
		return nil, err
	}
	return row.toModel(), nil
}
