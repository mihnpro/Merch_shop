package model

import (
	"errors"
	"testing"

	"github.com/google/uuid"

	"github.com/mihnpro/Merch_shop/services/order/internal/domain"
)

func sampleItems() []OrderItem {
	return []OrderItem{
		{ProductID: uuid.New(), ProductName: "Cap", Quantity: 2, PricePoints: 150},
		{ProductID: uuid.New(), ProductName: "Mug", Quantity: 1, PricePoints: 300},
	}
}

func TestNewOrder_ComputesTotalAndDefaults(t *testing.T) {
	userID := uuid.New()
	order, err := NewOrder(userID, sampleItems(), "Main St 1", nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if order.TotalPoints != 2*150+300 {
		t.Errorf("total = %d, want %d", order.TotalPoints, 2*150+300)
	}
	if order.Status != StatusPending {
		t.Errorf("status = %q, want pending", order.Status)
	}
	if order.CreatedAt.IsZero() || order.UpdatedAt.IsZero() {
		t.Error("timestamps must be set by the factory")
	}
	if order.ID == uuid.Nil {
		t.Error("order id must be generated")
	}
	for _, it := range order.Items {
		if it.ID == uuid.Nil || it.OrderID != order.ID {
			t.Errorf("item identity not wired: %+v", it)
		}
	}
}

func TestNewOrder_Invalid(t *testing.T) {
	cases := []struct {
		name    string
		items   []OrderItem
		address string
		want    error
	}{
		{"empty cart", nil, "Main St", domain.ErrEmptyCart},
		{"empty address", sampleItems(), "", domain.ErrInvalidInput},
		{"zero quantity", []OrderItem{{ProductID: uuid.New(), Quantity: 0, PricePoints: 10}}, "Main St", domain.ErrInvalidInput},
		{"nil product", []OrderItem{{ProductID: uuid.Nil, Quantity: 1, PricePoints: 10}}, "Main St", domain.ErrInvalidInput},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if _, err := NewOrder(uuid.New(), tc.items, tc.address, nil); !errors.Is(err, tc.want) {
				t.Errorf("err = %v, want %v", err, tc.want)
			}
		})
	}
}

func TestTransitionTo(t *testing.T) {
	order, _ := NewOrder(uuid.New(), sampleItems(), "Main St", nil)

	if err := order.TransitionTo(StatusConfirmed); err != nil {
		t.Fatalf("pending->confirmed should be allowed: %v", err)
	}
	if order.Status != StatusConfirmed {
		t.Errorf("status = %q, want confirmed", order.Status)
	}
	if err := order.TransitionTo(StatusDelivered); !errors.Is(err, domain.ErrInvalidStatusChange) {
		t.Errorf("err = %v, want ErrInvalidStatusChange", err)
	}
}
