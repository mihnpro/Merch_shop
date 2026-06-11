package repository

import (
	"context"

	"github.com/google/uuid"

	"github.com/mihnpro/Merch_shop/services/order/internal/domain/model"
)

type ListOrdersFilter struct {
	UserID    *uuid.UUID
	Status    string
	PageSize  int
	PageToken string
}

type OrderRepository interface {
	CreateOrder(ctx context.Context, order *model.Order, payload []byte) error
	GetOrderByID(ctx context.Context, id uuid.UUID) (*model.Order, error)
	UpdateOrderStatus(ctx context.Context, id uuid.UUID, status model.OrderStatus) error
	ListOrders(ctx context.Context, f ListOrdersFilter) ([]*model.Order, string, error)
}
