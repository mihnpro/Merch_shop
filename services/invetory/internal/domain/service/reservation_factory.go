package service

import (
	"github.com/google/uuid"

	"github.com/mihnpro/Merch_shop/services/invetory/internal/domain"
	"github.com/mihnpro/Merch_shop/services/invetory/internal/domain/model"
	vo "github.com/mihnpro/Merch_shop/services/invetory/internal/domain/valueobject"
)

type ReservationItemInput struct {
	ProductID uuid.UUID
	Qty       int
}

type ReserveInput struct {
	OrderID uuid.UUID
	Items   []ReservationItemInput
}

type ReservationFactory struct{}

func NewReservationFactory() *ReservationFactory { return &ReservationFactory{} }

func (f *ReservationFactory) Create(in ReserveInput) (*model.Reservation, error) {
	if len(in.Items) == 0 {
		return nil, domain.ErrEmptyReservation
	}
	items := make([]model.ReservationItem, 0, len(in.Items))
	for _, raw := range in.Items {
		qty, err := vo.NewQuantity(raw.Qty)
		if err != nil {
			return nil, err
		}
		items = append(items, model.ReservationItem{
			ProductID: raw.ProductID,
			Qty:       qty,
		})
	}
	return model.NewReservation(in.OrderID, items)
}
