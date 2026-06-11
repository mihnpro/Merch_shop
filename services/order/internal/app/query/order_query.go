package query

import (
	"context"

	"github.com/google/uuid"

	"github.com/mihnpro/Merch_shop/services/order/internal/app/dto"
	"github.com/mihnpro/Merch_shop/services/order/internal/domain"
	"github.com/mihnpro/Merch_shop/services/order/internal/domain/repository"
)

type OrderQueryService interface {
	GetOrder(ctx context.Context, in dto.GetOrderInput) (dto.OrderView, error)
	ListOrders(ctx context.Context, in dto.ListOrdersInput) ([]dto.OrderView, string, error)
}

type orderQueryService struct {
	orders repository.OrderRepository
}

func NewOrderQueryService(orders repository.OrderRepository) OrderQueryService {
	return &orderQueryService{orders: orders}
}

func (s *orderQueryService) GetOrder(ctx context.Context, in dto.GetOrderInput) (dto.OrderView, error) {
	id, err := uuid.Parse(in.OrderID)
	if err != nil {
		return dto.OrderView{}, domain.ErrInvalidInput
	}
	order, err := s.orders.GetOrderByID(ctx, id)
	if err != nil {
		return dto.OrderView{}, err
	}
	return dto.ToOrderView(order), nil
}

func (s *orderQueryService) ListOrders(ctx context.Context, in dto.ListOrdersInput) ([]dto.OrderView, string, error) {
	f := repository.ListOrdersFilter{
		Status:    in.Status,
		PageSize:  in.PageSize,
		PageToken: in.PageToken,
	}
	if in.UserID != "" {
		id, err := uuid.Parse(in.UserID)
		if err != nil {
			return nil, "", domain.ErrInvalidInput
		}
		f.UserID = &id
	}
	orders, nextToken, err := s.orders.ListOrders(ctx, f)
	if err != nil {
		return nil, "", err
	}
	views := make([]dto.OrderView, 0, len(orders))
	for _, o := range orders {
		views = append(views, dto.ToOrderView(o))
	}
	return views, nextToken, nil
}
