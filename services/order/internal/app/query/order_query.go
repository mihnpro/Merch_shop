package query

import (
	"context"
	"time"

	"github.com/google/uuid"

	"github.com/mihnpro/Merch_shop/services/order/internal/app/dto"
	"github.com/mihnpro/Merch_shop/services/order/internal/domain"
	"github.com/mihnpro/Merch_shop/services/order/internal/domain/repository"
)

type OrderQueryService interface {
	GetOrder(ctx context.Context, in dto.GetOrderInput) (dto.OrderView, error)
	ListOrders(ctx context.Context, in dto.ListOrdersInput) ([]dto.OrderView, string, error)
	GetAnalytics(ctx context.Context, period string) (dto.AnalyticsView, error)
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

func (s *orderQueryService) GetAnalytics(ctx context.Context, period string) (dto.AnalyticsView, error) {
	a, err := s.orders.GetAnalytics(ctx, analyticsSince(period))
	if err != nil {
		return dto.AnalyticsView{}, err
	}
	var avg int64
	if a.OrdersCount > 0 {
		avg = a.PointsSpent / int64(a.OrdersCount)
	}
	top := make([]dto.TopProductView, 0, len(a.TopProducts))
	for _, p := range a.TopProducts {
		top = append(top, dto.TopProductView{
			ProductID:   p.ProductID.String(),
			ProductName: p.ProductName,
			Quantity:    p.Quantity,
		})
	}
	return dto.AnalyticsView{
		Period:            normalizePeriod(period),
		OrdersCount:       a.OrdersCount,
		PointsSpent:       a.PointsSpent,
		AverageOrderValue: avg,
		TopProducts:       top,
	}, nil
}

func analyticsSince(period string) time.Time {
	now := time.Now()
	switch period {
	case "week":
		return now.AddDate(0, 0, -7)
	case "month":
		return now.AddDate(0, -1, 0)
	default:
		return now.AddDate(0, 0, -1)
	}
}

func normalizePeriod(period string) string {
	switch period {
	case "week", "month":
		return period
	default:
		return "day"
	}
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
