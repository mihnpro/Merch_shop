package service

import (
	"context"

	"github.com/mihnpro/Merch_shop/services/user_customer/internal/domain/model"
	"github.com/mihnpro/Merch_shop/services/user_customer/internal/domain/repository"
)


type PointsManager struct {
	balances repository.BalanceRepository
}

func NewPointsManager(balances repository.BalanceRepository) *PointsManager {
	return &PointsManager{balances: balances}
}

func (m *PointsManager) Credit(ctx context.Context, cmd repository.ApplyPointsCommand, amount int64) (model.PointsBalance, error) {
	return m.balances.Apply(ctx, cmd, func(b *model.PointsBalance) error {
		return b.Add(amount)
	})
}

func (m *PointsManager) Debit(ctx context.Context, cmd repository.ApplyPointsCommand, amount int64) (model.PointsBalance, error) {
	return m.balances.Apply(ctx, cmd, func(b *model.PointsBalance) error {
		return b.Deduct(amount)
	})
}
