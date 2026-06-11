package psql

import (
	"context"
	"database/sql"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"

	"github.com/mihnpro/Merch_shop/services/order/internal/domain"
	"github.com/mihnpro/Merch_shop/services/order/internal/domain/model"
	"github.com/mihnpro/Merch_shop/services/order/internal/domain/repository"
)

type orderRow struct {
	ID              uuid.UUID `db:"id"`
	UserID          uuid.UUID `db:"user_id"`
	TotalPoints     int64     `db:"total_points"`
	Status          string    `db:"status"`
	DeliveryAddress string    `db:"delivery_address"`
	CreatedAt       time.Time `db:"created_at"`
	UpdatedAt       time.Time `db:"updated_at"`
}

type orderItemRow struct {
	ID          uuid.UUID `db:"id"`
	OrderID     uuid.UUID `db:"order_id"`
	ProductID   uuid.UUID `db:"product_id"`
	ProductName string    `db:"product_name"`
	Quantity    int       `db:"quantity"`
	PricePoints int64     `db:"price_points"`
}


func insertOrder(ctx context.Context, exec sqlx.ExecerContext, order *model.Order) error {
	_, err := exec.ExecContext(ctx,
		`INSERT INTO orders (id, user_id, total_points, status, delivery_address, created_at, updated_at)
		 VALUES ($1, $2, $3, $4, $5, $6, $7)`,
		order.ID, order.UserID, order.TotalPoints, string(order.Status),
		order.DeliveryAddress, order.CreatedAt, order.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("insert order: %w", err)
	}
	return nil
}

func insertOrderItems(ctx context.Context, exec sqlx.ExecerContext, items []model.OrderItem) error {
	for _, item := range items {
		_, err := exec.ExecContext(ctx,
			`INSERT INTO order_items (id, order_id, product_id, product_name, quantity, price_points)
			 VALUES ($1, $2, $3, $4, $5, $6)`,
			item.ID, item.OrderID, item.ProductID, item.ProductName, item.Quantity, item.PricePoints,
		)
		if err != nil {
			return fmt.Errorf("insert order item: %w", err)
		}
	}
	return nil
}

func getOrderRow(ctx context.Context, q sqlx.QueryerContext, id uuid.UUID) (*orderRow, error) {
	var row orderRow
	err := sqlx.GetContext(ctx, q, &row,
		`SELECT id, user_id, total_points, status, delivery_address, created_at, updated_at
		 FROM orders WHERE id = $1`, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, domain.ErrOrderNotFound
		}
		return nil, err
	}
	return &row, nil
}

func selectOrderItems(ctx context.Context, q sqlx.QueryerContext, orderID uuid.UUID) ([]orderItemRow, error) {
	var itemRows []orderItemRow
	if err := sqlx.SelectContext(ctx, q, &itemRows,
		`SELECT id, order_id, product_id, product_name, quantity, price_points
		 FROM order_items WHERE order_id = $1`, orderID); err != nil {
		return nil, err
	}
	return itemRows, nil
}

func updateOrderStatus(ctx context.Context, exec sqlx.ExecerContext, id uuid.UUID, status model.OrderStatus) error {
	res, err := exec.ExecContext(ctx,
		`UPDATE orders SET status = $1, updated_at = NOW() WHERE id = $2`,
		string(status), id)
	if err != nil {
		return err
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return domain.ErrOrderNotFound
	}
	return nil
}

func selectOrdersPage(ctx context.Context, db *sqlx.DB, f repository.ListOrdersFilter, pageSize int) ([]orderRow, error) {
	args := []interface{}{}
	where := "1=1"
	idx := 1

	if f.UserID != nil {
		where += fmt.Sprintf(" AND o.user_id = $%d", idx)
		args = append(args, *f.UserID)
		idx++
	}
	if f.Status != "" {
		where += fmt.Sprintf(" AND o.status = $%d", idx)
		args = append(args, f.Status)
		idx++
	}
	if f.PageToken != "" {
		cur, err := decodeOrderCursor(f.PageToken)
		if err == nil {
			where += fmt.Sprintf(" AND (o.created_at, o.id) < ($%d, $%d)", idx, idx+1)
			args = append(args, cur.CreatedAt, cur.ID)
			idx += 2
		}
	}

	args = append(args, pageSize+1)
	q := fmt.Sprintf(
		`SELECT id, user_id, total_points, status, delivery_address, created_at, updated_at
		 FROM orders o
		 WHERE %s
		 ORDER BY o.created_at DESC, o.id DESC
		 LIMIT $%d`, where, idx)

	var rows []orderRow
	if err := db.SelectContext(ctx, &rows, q, args...); err != nil {
		return nil, err
	}
	return rows, nil
}

func selectOrderItemsByIDs(ctx context.Context, db *sqlx.DB, ids []uuid.UUID) ([]orderItemRow, error) {
	q, args, err := sqlx.In(
		`SELECT id, order_id, product_id, product_name, quantity, price_points
		 FROM order_items WHERE order_id IN (?)`, ids)
	if err != nil {
		return nil, err
	}
	q = db.Rebind(q)
	var itemRows []orderItemRow
	if err := db.SelectContext(ctx, &itemRows, q, args...); err != nil {
		return nil, err
	}
	return itemRows, nil
}


type orderCursor struct {
	CreatedAt time.Time `json:"ca"`
	ID        uuid.UUID `json:"id"`
}

func encodeOrderCursor(c orderCursor) string {
	b, _ := json.Marshal(c)
	return base64.StdEncoding.EncodeToString(b)
}

func decodeOrderCursor(s string) (orderCursor, error) {
	b, err := base64.StdEncoding.DecodeString(s)
	if err != nil {
		return orderCursor{}, err
	}
	var c orderCursor
	return c, json.Unmarshal(b, &c)
}

func toOrderModel(row *orderRow, itemRows []orderItemRow) *model.Order {
	items := make([]model.OrderItem, 0, len(itemRows))
	for _, it := range itemRows {
		items = append(items, model.OrderItem{
			ID:          it.ID,
			OrderID:     it.OrderID,
			ProductID:   it.ProductID,
			ProductName: it.ProductName,
			Quantity:    it.Quantity,
			PricePoints: it.PricePoints,
		})
	}
	return &model.Order{
		ID:              row.ID,
		UserID:          row.UserID,
		TotalPoints:     row.TotalPoints,
		Status:          model.OrderStatus(row.Status),
		DeliveryAddress: row.DeliveryAddress,
		Items:           items,
		CreatedAt:       row.CreatedAt,
		UpdatedAt:       row.UpdatedAt,
	}
}
