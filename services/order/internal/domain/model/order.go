package model

import (
	"time"

	"github.com/google/uuid"

	"github.com/mihnpro/Merch_shop/services/order/internal/domain"
)

type OrderStatus string

const (
	StatusPending       OrderStatus = "pending"
	StatusConfirmed     OrderStatus = "confirmed"
	StatusReadyToPickup OrderStatus = "ready_to_pickup"
	StatusDelivered     OrderStatus = "delivered"
	StatusCancelled     OrderStatus = "cancelled"
)

var validTransitions = map[OrderStatus][]OrderStatus{
	StatusPending:       {StatusConfirmed, StatusCancelled},
	StatusConfirmed:     {StatusReadyToPickup, StatusCancelled},
	StatusReadyToPickup: {StatusDelivered},
}

func (s OrderStatus) CanTransitionTo(next OrderStatus) bool {
	allowed, ok := validTransitions[s]
	if !ok {
		return false
	}
	for _, a := range allowed {
		if a == next {
			return true
		}
	}
	return false
}

type Order struct {
	ID              uuid.UUID
	UserID          uuid.UUID
	Items           []OrderItem
	TotalPoints     int64
	Status          OrderStatus
	Note            *string
	CancelReason    *string
	DeliveryAddress string
	CreatedAt       time.Time
	UpdatedAt       time.Time
}

type OrderItem struct {
	ID          uuid.UUID
	OrderID     uuid.UUID
	ProductID   uuid.UUID
	ProductName string
	Quantity    int
	PricePoints int64
}

func NewOrder(userID uuid.UUID, items []OrderItem, deliveryAddress string, note *string) (*Order, error) {
	if deliveryAddress == "" {
		return nil, domain.ErrInvalidInput
	}
	if len(items) == 0 {
		return nil, domain.ErrEmptyCart
	}

	orderID := uuid.New()
	now := time.Now().UTC()

	var total int64
	for i := range items {
		it := &items[i]
		if it.Quantity <= 0 || it.PricePoints <= 0 || it.ProductID == uuid.Nil {
			return nil, domain.ErrInvalidInput
		}
		if it.ID == uuid.Nil {
			it.ID = uuid.New()
		}
		it.OrderID = orderID
		total += it.PricePoints * int64(it.Quantity)
	}

	return &Order{
		ID:              orderID,
		UserID:          userID,
		Items:           items,
		TotalPoints:     total,
		Status:          StatusPending,
		Note:            note,
		DeliveryAddress: deliveryAddress,
		CreatedAt:       now,
		UpdatedAt:       now,
	}, nil
}

func (o *Order) TransitionTo(next OrderStatus) error {
	if !o.Status.CanTransitionTo(next) {
		return domain.ErrInvalidStatusChange
	}
	o.Status = next
	o.UpdatedAt = time.Now().UTC()
	return nil
}
