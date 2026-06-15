package service

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/mihnpro/Merch_shop/services/user_customer/internal/domain"
	"github.com/mihnpro/Merch_shop/services/user_customer/internal/domain/model"
	"github.com/mihnpro/Merch_shop/services/user_customer/internal/domain/repository"
)

type fakeBalanceRepo struct {
	balances map[uuid.UUID]*model.PointsBalance
	applyErr error
}

func newFakeBalanceRepo() *fakeBalanceRepo {
	return &fakeBalanceRepo{balances: make(map[uuid.UUID]*model.PointsBalance)}
}

func (f *fakeBalanceRepo) EnsureBalance(_ context.Context, userID uuid.UUID) error {
	if _, ok := f.balances[userID]; !ok {
		f.balances[userID] = &model.PointsBalance{UserID: userID, Points: 0, UpdatedAt: time.Now()}
	}
	return nil
}

func (f *fakeBalanceRepo) GetBalance(_ context.Context, userID uuid.UUID) (model.PointsBalance, error) {
	b, ok := f.balances[userID]
	if !ok {
		return model.PointsBalance{}, nil
	}
	return *b, nil
}

func (f *fakeBalanceRepo) Apply(_ context.Context, cmd repository.ApplyPointsCommand, mutate func(*model.PointsBalance) error) (model.PointsBalance, error) {
	if f.applyErr != nil {
		return model.PointsBalance{}, f.applyErr
	}
	b, ok := f.balances[cmd.UserID]
	if !ok {
		b = &model.PointsBalance{UserID: cmd.UserID, Points: 0, UpdatedAt: time.Now()}
		f.balances[cmd.UserID] = b
	}
	if err := mutate(b); err != nil {
		return model.PointsBalance{}, err
	}
	return *b, nil
}

func (f *fakeBalanceRepo) GetTransactions(_ context.Context, _ uuid.UUID, _ int, _ string) ([]model.PointsTransaction, string, error) {
	return nil, "", nil
}

func TestPointsManager_Credit(t *testing.T) {
	t.Run("happy path", func(t *testing.T) {
		fb := newFakeBalanceRepo()
		pm := NewPointsManager(fb)
		uid := uuid.New()
		fb.balances[uid] = &model.PointsBalance{UserID: uid, Points: 100, UpdatedAt: time.Now()}

		cmd := repository.ApplyPointsCommand{UserID: uid, OperationID: uuid.New(), Reason: "bonus"}
		bal, err := pm.Credit(context.Background(), cmd, 50)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if bal.Points != 150 {
			t.Errorf("points = %d, want 150", bal.Points)
		}
	})

	t.Run("credit zero balance", func(t *testing.T) {
		fb := newFakeBalanceRepo()
		pm := NewPointsManager(fb)
		uid := uuid.New()

		cmd := repository.ApplyPointsCommand{UserID: uid, OperationID: uuid.New(), Reason: "first"}
		bal, err := pm.Credit(context.Background(), cmd, 100)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if bal.Points != 100 {
			t.Errorf("points = %d, want 100", bal.Points)
		}
	})

	t.Run("repo error", func(t *testing.T) {
		fb := newFakeBalanceRepo()
		fb.applyErr = errBoom
		pm := NewPointsManager(fb)

		cmd := repository.ApplyPointsCommand{UserID: uuid.New(), OperationID: uuid.New(), Reason: "test"}
		_, err := pm.Credit(context.Background(), cmd, 50)
		if err != errBoom {
			t.Errorf("err = %v, want errBoom", err)
		}
	})
}

func TestPointsManager_Debit(t *testing.T) {
	t.Run("happy path", func(t *testing.T) {
		fb := newFakeBalanceRepo()
		pm := NewPointsManager(fb)
		uid := uuid.New()
		fb.balances[uid] = &model.PointsBalance{UserID: uid, Points: 200, UpdatedAt: time.Now()}

		cmd := repository.ApplyPointsCommand{UserID: uid, OperationID: uuid.New(), Reason: "order"}
		bal, err := pm.Debit(context.Background(), cmd, 80)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if bal.Points != 120 {
			t.Errorf("points = %d, want 120", bal.Points)
		}
	})

	t.Run("insufficient balance", func(t *testing.T) {
		fb := newFakeBalanceRepo()
		pm := NewPointsManager(fb)
		uid := uuid.New()
		fb.balances[uid] = &model.PointsBalance{UserID: uid, Points: 30, UpdatedAt: time.Now()}

		cmd := repository.ApplyPointsCommand{UserID: uid, OperationID: uuid.New(), Reason: "order"}
		_, err := pm.Debit(context.Background(), cmd, 100)
		if err != domain.ErrInsufficientBalance {
			t.Errorf("err = %v, want ErrInsufficientBalance", err)
		}
	})

	t.Run("zero amount", func(t *testing.T) {
		fb := newFakeBalanceRepo()
		pm := NewPointsManager(fb)
		uid := uuid.New()
		fb.balances[uid] = &model.PointsBalance{UserID: uid, Points: 100, UpdatedAt: time.Now()}

		cmd := repository.ApplyPointsCommand{UserID: uid, OperationID: uuid.New(), Reason: "test"}
		_, err := pm.Debit(context.Background(), cmd, 0)
		if err != domain.ErrInvalidInput {
			t.Errorf("err = %v, want ErrInvalidInput", err)
		}
	})

	t.Run("repo error", func(t *testing.T) {
		fb := newFakeBalanceRepo()
		fb.applyErr = errBoom
		pm := NewPointsManager(fb)

		cmd := repository.ApplyPointsCommand{UserID: uuid.New(), OperationID: uuid.New(), Reason: "test"}
		_, err := pm.Debit(context.Background(), cmd, 50)
		if err != errBoom {
			t.Errorf("err = %v, want errBoom", err)
		}
	})
}
