package model

import (
	"time"

	"github.com/google/uuid"

	"github.com/mihnpro/Merch_shop/services/invetory/internal/domain"
	vo "github.com/mihnpro/Merch_shop/services/invetory/internal/domain/valueobject"
)

type ReservationItem struct {
	ProductID uuid.UUID
	Qty       vo.Quantity
}

type Reservation struct {
	ID             uuid.UUID
	OrderID        uuid.UUID
	Items          []ReservationItem
	ReleasedReason string
	ReleasedAt     *time.Time
	CreatedAt      time.Time
	UpdatedAt      time.Time
}

func NewReservation(orderID uuid.UUID, items []ReservationItem) (*Reservation, error) {
	r := &Reservation{
		OrderID: orderID,
		Items:   items,
	}
	if err := r.Validate(); err != nil {
		return nil, err
	}
	return r, nil
}

func (r *Reservation) Validate() error {
	if r.OrderID == uuid.Nil {
		return domain.ErrInvalidOrderID
	}
	if len(r.Items) == 0 {
		return domain.ErrEmptyReservation
	}
	for _, it := range r.Items {
		if it.ProductID == uuid.Nil {
			return domain.ErrInvalidProductID
		}
		if it.Qty.Int() < 1 {
			return domain.ErrInvalidQuantity
		}
	}
	return nil
}

func (r *Reservation) IsReleased() bool {
	return r.ReleasedAt != nil
}

func (r *Reservation) Release(reason string) error {
	if r.IsReleased() {
		return domain.ErrAlreadyReleased
	}
	r.ReleasedReason = reason
	return nil
}
