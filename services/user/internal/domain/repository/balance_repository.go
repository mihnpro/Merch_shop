package repository

import (
	"context"

	"github.com/google/uuid"

	"github.com/mihnpro/Merch_shop/services/user_customer/internal/domain/model"
)

type ApplyPointsCommand struct {
	UserID      uuid.UUID
	OperationID uuid.UUID
	OrderID     *uuid.UUID
	Reason      string
}

type BalanceRepository interface {
	EnsureBalance(ctx context.Context, userID uuid.UUID) error
	GetBalance(ctx context.Context, userID uuid.UUID) (model.PointsBalance, error)
	Apply(ctx context.Context, cmd ApplyPointsCommand, mutate func(*model.PointsBalance) error) (model.PointsBalance, error)
	GetTransactions(ctx context.Context, userID uuid.UUID, limit int, cursor string) ([]model.PointsTransaction, string, error)
}
