package model

import (
	"time"

	"github.com/google/uuid"

	"github.com/mihnpro/Merch_shop/services/invetory/internal/domain"
)

type Stock struct {
	ProductID uuid.UUID
	Available int
	Version   int
	CreatedAt time.Time
	UpdatedAt time.Time
}

func NewStock(productID uuid.UUID) (*Stock, error) {
	s := &Stock{
		ProductID: productID,
		Available: 0,
		Version:   1,
	}
	if err := s.Validate(); err != nil {
		return nil, err
	}
	return s, nil
}

func (s *Stock) Validate() error {
	if s.ProductID == uuid.Nil {
		return domain.ErrInvalidProductID
	}
	if s.Available < 0 {
		return domain.ErrInsufficientStock
	}
	return nil
}

func (s *Stock) Adjust(delta int) error {
	if s.Available+delta < 0 {
		return domain.ErrInsufficientStock
	}
	s.Available += delta
	return nil
}

func (s *Stock) Reserve(qty int) error {
	if qty > s.Available {
		return domain.ErrInsufficientStock
	}
	s.Available -= qty
	return nil
}

func (s *Stock) Release(qty int) {
	s.Available += qty
}

func (s *Stock) CanReserve(qty int) bool {
	return s.Available >= qty
}
