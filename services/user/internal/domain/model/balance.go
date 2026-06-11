package model

import (
	"time"

	"github.com/google/uuid"

	"github.com/mihnpro/Merch_shop/services/user_customer/internal/domain"
)

type PointsBalance struct {
	UserID    uuid.UUID
	Points    int64
	UpdatedAt time.Time
}


func (b *PointsBalance) Add(amount int64) error {
	if amount <= 0 {
		return domain.ErrInvalidInput
	}
	b.Points += amount
	return nil
}

func (b *PointsBalance) Deduct(amount int64) error {
	if amount <= 0 {
		return domain.ErrInvalidInput
	}
	if b.Points < amount {
		return domain.ErrInsufficientBalance
	}
	b.Points -= amount
	return nil
}

type PointsTransaction struct {
	ID          uuid.UUID
	UserID      uuid.UUID
	OperationID uuid.UUID
	OrderID     *uuid.UUID
	Amount      int64
	Reason      string
	CreatedAt   time.Time
}
