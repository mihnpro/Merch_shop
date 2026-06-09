package repository

import (
	"context"

	"github.com/google/uuid"

	"github.com/mihnpro/Merch_shop/services/invetory/internal/domain/model"
)

type ReservationRepository interface {

	GetByOrderID(ctx context.Context, orderID uuid.UUID) (*model.Reservation, error)

	Reserve(ctx context.Context, r *model.Reservation) (*model.Reservation, error)

	Release(ctx context.Context, orderID uuid.UUID, reason string) (*model.Reservation, error)
}
