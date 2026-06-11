package model

import (
	"time"

	"github.com/google/uuid"

	vo "github.com/mihnpro/Merch_shop/services/cart/internal/domain/valueobject"
)

type CartItem struct {
	ID               uuid.UUID
	CartID           uuid.UUID
	ProductID        uuid.UUID
	Quantity         vo.Quantity
	PriceAtAdd       int
	ProductNameAtAdd string
	CreatedAt        time.Time
	UpdatedAt        time.Time
}
