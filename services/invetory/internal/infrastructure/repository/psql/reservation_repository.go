package psql

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"

	"github.com/mihnpro/Merch_shop/services/invetory/internal/domain"
	"github.com/mihnpro/Merch_shop/services/invetory/internal/domain/model"
	"github.com/mihnpro/Merch_shop/services/invetory/internal/domain/repository"
	vo "github.com/mihnpro/Merch_shop/services/invetory/internal/domain/valueobject"
)

type reservationRepository struct {
	db *sqlx.DB
}

func NewReservationRepository(db *sqlx.DB) repository.ReservationRepository {
	return &reservationRepository{db: db}
}

type itemJSON struct {
	ProductID uuid.UUID `json:"product_id"`
	Qty       int       `json:"qty"`
}

type reservationRow struct {
	ID             uuid.UUID      `db:"id"`
	OrderID        uuid.UUID      `db:"order_id"`
	Items          []byte         `db:"items"`
	ReleasedReason sql.NullString `db:"released_reason"`
	ReleasedAt     sql.NullTime   `db:"released_at"`
	CreatedAt      time.Time      `db:"created_at"`
	UpdatedAt      time.Time      `db:"updated_at"`
}

func (r *reservationRow) toModel() (*model.Reservation, error) {
	var raw []itemJSON
	if err := json.Unmarshal(r.Items, &raw); err != nil {
		return nil, err
	}
	items := make([]model.ReservationItem, 0, len(raw))
	for _, it := range raw {
		items = append(items, model.ReservationItem{
			ProductID: it.ProductID,
			Qty:       vo.NewQuantityFromStored(it.Qty),
		})
	}

	res := &model.Reservation{
		ID:        r.ID,
		OrderID:   r.OrderID,
		Items:     items,
		CreatedAt: r.CreatedAt,
		UpdatedAt: r.UpdatedAt,
	}
	if r.ReleasedReason.Valid {
		res.ReleasedReason = r.ReleasedReason.String
	}
	if r.ReleasedAt.Valid {
		t := r.ReleasedAt.Time
		res.ReleasedAt = &t
	}
	return res, nil
}

func marshalItems(items []model.ReservationItem) ([]byte, error) {
	raw := make([]itemJSON, 0, len(items))
	for _, it := range items {
		raw = append(raw, itemJSON{ProductID: it.ProductID, Qty: it.Qty.Int()})
	}
	return json.Marshal(raw)
}

const selectReservation = `
	SELECT id, order_id, items, released_reason, released_at, created_at, updated_at
	FROM reservations
	WHERE order_id = $1`

func (r *reservationRepository) GetByOrderID(ctx context.Context, orderID uuid.UUID) (*model.Reservation, error) {
	var row reservationRow
	if err := r.db.GetContext(ctx, &row, selectReservation, orderID); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, domain.ErrReservationNotFound
		}
		return nil, err
	}
	return row.toModel()
}

const insertReservation = `
	INSERT INTO reservations (order_id, items)
	VALUES ($1, $2)
	ON CONFLICT (order_id) DO NOTHING
	RETURNING id, created_at, updated_at`

const decrementStock = `
	UPDATE stock
	SET available = available - $1, version = version + 1, updated_at = NOW()
	WHERE product_id = $2 AND available >= $1`

func (r *reservationRepository) Reserve(ctx context.Context, res *model.Reservation) (*model.Reservation, error) {
	itemsJSON, err := marshalItems(res.Items)
	if err != nil {
		return nil, err
	}

	tx, err := r.db.BeginTxx(ctx, nil)
	if err != nil {
		return nil, err
	}
	defer func() { _ = tx.Rollback() }()

	err = tx.QueryRowContext(ctx, insertReservation, res.OrderID, itemsJSON).
		Scan(&res.ID, &res.CreatedAt, &res.UpdatedAt)

	if errors.Is(err, sql.ErrNoRows) {
		existing, gerr := getReservationTx(ctx, tx, res.OrderID)
		if gerr != nil {
			return nil, gerr
		}
		if cerr := tx.Commit(); cerr != nil {
			return nil, cerr
		}
		return existing, nil
	}
	if err != nil {
		return nil, err
	}

	for _, it := range res.Items {
		result, err := tx.ExecContext(ctx, decrementStock, it.Qty.Int(), it.ProductID)
		if err != nil {
			return nil, err
		}
		affected, err := result.RowsAffected()
		if err != nil {
			return nil, err
		}
		if affected == 0 {
			return nil, domain.ErrInsufficientStock
		}
	}

	if err := tx.Commit(); err != nil {
		return nil, err
	}
	return res, nil
}

const releaseReservation = `
	UPDATE reservations
	SET released_at = NOW(), released_reason = $2
	WHERE order_id = $1 AND released_at IS NULL
	RETURNING id, order_id, items, released_reason, released_at, created_at, updated_at`

const restoreStock = `
	UPDATE stock
	SET available = available + $1, version = version + 1, updated_at = NOW()
	WHERE product_id = $2`

func (r *reservationRepository) Release(ctx context.Context, orderID uuid.UUID, reason string) (*model.Reservation, error) {
	tx, err := r.db.BeginTxx(ctx, nil)
	if err != nil {
		return nil, err
	}
	defer func() { _ = tx.Rollback() }()

	var row reservationRow
	err = tx.QueryRowContext(ctx, releaseReservation, orderID, nullReason(reason)).Scan(
		&row.ID, &row.OrderID, &row.Items, &row.ReleasedReason, &row.ReleasedAt, &row.CreatedAt, &row.UpdatedAt,
	)
	
	if errors.Is(err, sql.ErrNoRows) {
		existing, gerr := getReservationTx(ctx, tx, orderID)
		if gerr != nil {
			return nil, gerr
		}
		if cerr := tx.Commit(); cerr != nil {
			return nil, cerr
		}
		return existing, nil
	}
	if err != nil {
		return nil, err
	}

	released, err := row.toModel()
	if err != nil {
		return nil, err
	}

	for _, it := range released.Items {
		if _, err := tx.ExecContext(ctx, restoreStock, it.Qty.Int(), it.ProductID); err != nil {
			return nil, err
		}
	}

	if err := tx.Commit(); err != nil {
		return nil, err
	}
	return released, nil
}

func getReservationTx(ctx context.Context, tx *sqlx.Tx, orderID uuid.UUID) (*model.Reservation, error) {
	var row reservationRow
	if err := tx.GetContext(ctx, &row, selectReservation, orderID); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, domain.ErrReservationNotFound
		}
		return nil, err
	}
	return row.toModel()
}

func nullReason(reason string) sql.NullString {
	reason = strings.TrimSpace(reason)
	if reason == "" {
		return sql.NullString{}
	}
	return sql.NullString{String: reason, Valid: true}
}
