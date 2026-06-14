package psql

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"

	"github.com/mihnpro/Merch_shop/services/order/internal/domain/events"
	"github.com/mihnpro/Merch_shop/services/order/internal/domain/model"
	"github.com/mihnpro/Merch_shop/services/order/internal/domain/repository"
)



type outboxWriter interface {
	Add(ctx context.Context, exec sqlx.ExecerContext, eventType string, payload []byte) error
}

type orderRepository struct {
	db     *sqlx.DB
	outbox outboxWriter
}

func NewOrderRepository(db *sqlx.DB, outbox outboxWriter) repository.OrderRepository {
	return &orderRepository{db: db, outbox: outbox}
}

func (r *orderRepository) withTx(ctx context.Context, fn func(tx *sqlx.Tx) error) (err error) {
	tx, err := r.db.BeginTxx(ctx, nil)
	if err != nil {
		return err
	}
	defer func() {
		if p := recover(); p != nil {
			_ = tx.Rollback()
			panic(p)
		} else if err != nil {
			_ = tx.Rollback()
		}
	}()

	if err = fn(tx); err != nil {
		return err
	}
	return tx.Commit()
}

func (r *orderRepository) CreateOrder(ctx context.Context, order *model.Order, payload []byte) error {
	return r.withTx(ctx, func(tx *sqlx.Tx) error {
		if err := insertOrder(ctx, tx, order); err != nil {
			return err
		}
		if err := insertOrderItems(ctx, tx, order.Items); err != nil {
			return err
		}
		return r.outbox.Add(ctx, tx, events.TypeOrderCreated, payload)
	})
}

func (r *orderRepository) GetOrderByID(ctx context.Context, id uuid.UUID) (*model.Order, error) {
	row, err := getOrderRow(ctx, r.db, id)
	if err != nil {
		return nil, err
	}
	itemRows, err := selectOrderItems(ctx, r.db, id)
	if err != nil {
		return nil, err
	}
	return toOrderModel(row, itemRows), nil
}

func (r *orderRepository) UpdateOrderStatus(ctx context.Context, id uuid.UUID, status model.OrderStatus) error {
	return updateOrderStatus(ctx, r.db, id, status)
}

func (r *orderRepository) GetAnalytics(ctx context.Context, since time.Time) (repository.Analytics, error) {
	return selectAnalytics(ctx, r.db, since)
}

func (r *orderRepository) ListOrders(ctx context.Context, f repository.ListOrdersFilter) ([]*model.Order, string, error) {
	pageSize := f.PageSize
	if pageSize <= 0 || pageSize > 100 {
		pageSize = 20
	}

	rows, err := selectOrdersPage(ctx, r.db, f, pageSize)
	if err != nil {
		return nil, "", err
	}

	var nextToken string
	if len(rows) > pageSize {
		rows = rows[:pageSize]
		last := rows[len(rows)-1]
		nextToken = encodeOrderCursor(orderCursor{CreatedAt: last.CreatedAt, ID: last.ID})
	}

	if len(rows) == 0 {
		return []*model.Order{}, nextToken, nil
	}

	ids := make([]uuid.UUID, 0, len(rows))
	for _, row := range rows {
		ids = append(ids, row.ID)
	}
	itemRows, err := selectOrderItemsByIDs(ctx, r.db, ids)
	if err != nil {
		return nil, "", err
	}

	itemsByOrder := make(map[uuid.UUID][]orderItemRow, len(rows))
	for _, it := range itemRows {
		itemsByOrder[it.OrderID] = append(itemsByOrder[it.OrderID], it)
	}

	orders := make([]*model.Order, 0, len(rows))
	for _, row := range rows {
		orders = append(orders, toOrderModel(&row, itemsByOrder[row.ID]))
	}
	return orders, nextToken, nil
}
