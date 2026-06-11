package service

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/google/uuid"

	"github.com/mihnpro/Merch_shop/services/order/internal/app/dto"
	"github.com/mihnpro/Merch_shop/services/order/internal/app/port"
	"github.com/mihnpro/Merch_shop/services/order/internal/domain"
	"github.com/mihnpro/Merch_shop/services/order/internal/domain/events"
	"github.com/mihnpro/Merch_shop/services/order/internal/domain/model"
	"github.com/mihnpro/Merch_shop/services/order/internal/domain/repository"
)

type OrderService interface {
	CreateOrder(ctx context.Context, in dto.CreateOrderInput) (dto.OrderView, error)
	UpdateOrderStatus(ctx context.Context, in dto.UpdateOrderStatusInput) (dto.OrderView, error)
}

type orderService struct {
	orders    repository.OrderRepository
	cart      port.CartClient
	user      port.UserClient
	inventory port.InventoryClient
}

func NewOrderService(
	orders repository.OrderRepository,
	cart port.CartClient,
	user port.UserClient,
	inventory port.InventoryClient,
) OrderService {
	return &orderService{orders: orders, cart: cart, user: user, inventory: inventory}
}

func (s *orderService) CreateOrder(ctx context.Context, in dto.CreateOrderInput) (dto.OrderView, error) {
	userID, err := uuid.Parse(in.UserID)
	if err != nil {
		return dto.OrderView{}, domain.ErrInvalidInput
	}

	items, err := s.cart.GetCart(ctx, in.UserID)
	if err != nil {
		return dto.OrderView{}, fmt.Errorf("get cart: %w", err)
	}

	orderItems := make([]model.OrderItem, 0, len(items))
	for _, ci := range items {
		pid, err := uuid.Parse(ci.ProductID)
		if err != nil {
			return dto.OrderView{}, domain.ErrInvalidInput
		}
		orderItems = append(orderItems, model.OrderItem{
			ProductID:   pid,
			ProductName: ci.ProductName,
			Quantity:    ci.Quantity,
			PricePoints: ci.PricePoints,
		})
	}

	order, err := model.NewOrder(userID, orderItems, in.DeliveryAddress)
	if err != nil {
		return dto.OrderView{}, err
	}

	balance, err := s.user.GetBalance(ctx, in.UserID)
	if err != nil {
		return dto.OrderView{}, fmt.Errorf("get balance: %w", err)
	}
	if balance < order.TotalPoints {
		return dto.OrderView{}, domain.ErrInsufficientBalance
	}

	opID := pointsOperationID(order.ID, "deduct")
	if err := s.user.DeductPoints(ctx, in.UserID, order.TotalPoints, opID, order.ID.String(), "order"); err != nil {
		return dto.OrderView{}, fmt.Errorf("deduct points: %w", err)
	}

	payload, err := json.Marshal(orderCreatedEvent(order))
	if err != nil {
		_ = s.user.AddPoints(ctx, in.UserID, order.TotalPoints, pointsOperationID(order.ID, "refund-marshal"), order.ID.String(), "order cancelled: internal error")
		return dto.OrderView{}, err
	}
	if err := s.orders.CreateOrder(ctx, order, payload); err != nil {
		_ = s.user.AddPoints(ctx, in.UserID, order.TotalPoints, pointsOperationID(order.ID, "refund-save"), order.ID.String(), "order cancelled: save error")
		return dto.OrderView{}, fmt.Errorf("save order: %w", err)
	}

	_ = s.cart.ClearCart(ctx, in.UserID)

	return dto.ToOrderView(order), nil
}


var opNamespace = uuid.MustParse("6f3c9b2e-7a1d-4c8e-9b5a-2d4f6a8c0e11")

func pointsOperationID(orderID uuid.UUID, kind string) string {
	return uuid.NewSHA1(opNamespace, []byte(orderID.String()+":"+kind)).String()
}

func orderCreatedEvent(order *model.Order) events.OrderCreated {
	evItems := make([]events.OrderCreatedItem, 0, len(order.Items))
	for _, it := range order.Items {
		evItems = append(evItems, events.OrderCreatedItem{
			ProductID: it.ProductID.String(),
			Quantity:  it.Quantity,
		})
	}
	return events.OrderCreated{
		OrderID: order.ID.String(),
		UserID:  order.UserID.String(),
		Items:   evItems,
	}
}

func (s *orderService) UpdateOrderStatus(ctx context.Context, in dto.UpdateOrderStatusInput) (dto.OrderView, error) {
	id, err := uuid.Parse(in.OrderID)
	if err != nil {
		return dto.OrderView{}, domain.ErrInvalidInput
	}

	order, err := s.orders.GetOrderByID(ctx, id)
	if err != nil {
		return dto.OrderView{}, err
	}

	newStatus := model.OrderStatus(in.NewStatus)

	if order.Status == newStatus {
		return dto.ToOrderView(order), nil
	}

	if err := order.TransitionTo(newStatus); err != nil {
		return dto.OrderView{}, err
	}

	if err := s.orders.UpdateOrderStatus(ctx, id, newStatus); err != nil {
		return dto.OrderView{}, err
	}
	
	if newStatus == model.StatusCancelled {
		opID := pointsOperationID(id, "refund")
		_ = s.user.AddPoints(ctx, order.UserID.String(), order.TotalPoints, opID, id.String(), "order cancelled")
		_ = s.inventory.ReleaseReserve(ctx, id.String(), in.Reason)
	}

	return dto.ToOrderView(order), nil
}
