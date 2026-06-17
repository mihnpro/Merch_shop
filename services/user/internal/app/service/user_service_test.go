package service

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"

	"github.com/mihnpro/Merch_shop/services/user_customer/internal/app/dto"
	"github.com/mihnpro/Merch_shop/services/user_customer/internal/domain"
	"github.com/mihnpro/Merch_shop/services/user_customer/internal/domain/model"
)

func TestUserService_GetUser(t *testing.T) {
	t.Run("invalid UUID", func(t *testing.T) {
		h := newHarness()
		_, err := h.userSvc.GetUser(context.Background(), "not-a-uuid")
		assert.ErrorIs(t, err, domain.ErrInvalidInput)
	})

	t.Run("user not found", func(t *testing.T) {
		h := newHarness()
		_, err := h.userSvc.GetUser(context.Background(), uuid.New().String())
		assert.ErrorIs(t, err, domain.ErrUserNotFound)
	})

	t.Run("happy path", func(t *testing.T) {
		h := newHarness()
		u := seedUser(h, "alice")
		view, err := h.userSvc.GetUser(context.Background(), u.ID.String())
		assert.NoError(t, err)
		assert.Equal(t, "alice", view.Login)
		assert.Equal(t, u.ID.String(), view.ID)
	})
}

func TestUserService_UpdateProfile(t *testing.T) {
	t.Run("invalid UUID", func(t *testing.T) {
		h := newHarness()
		_, err := h.userSvc.UpdateProfile(context.Background(), dto.UpdateProfileInput{UserID: "bad"})
		assert.ErrorIs(t, err, domain.ErrInvalidInput)
	})

	t.Run("empty first name", func(t *testing.T) {
		h := newHarness()
		_, err := h.userSvc.UpdateProfile(context.Background(), dto.UpdateProfileInput{
			UserID: uuid.New().String(), FirstName: "  ", LastName: "Doe", Email: "a@b.com",
		})
		assert.ErrorIs(t, err, domain.ErrEmptyFirstName)
	})

	t.Run("empty last name", func(t *testing.T) {
		h := newHarness()
		_, err := h.userSvc.UpdateProfile(context.Background(), dto.UpdateProfileInput{
			UserID: uuid.New().String(), FirstName: "John", LastName: "  ", Email: "a@b.com",
		})
		assert.ErrorIs(t, err, domain.ErrEmptyLastName)
	})

	t.Run("invalid email", func(t *testing.T) {
		h := newHarness()
		_, err := h.userSvc.UpdateProfile(context.Background(), dto.UpdateProfileInput{
			UserID: uuid.New().String(), FirstName: "John", LastName: "Doe", Email: "bad",
		})
		assert.ErrorIs(t, err, domain.ErrInvalidEmailFormat)
	})

	t.Run("invalid phone", func(t *testing.T) {
		h := newHarness()
		_, err := h.userSvc.UpdateProfile(context.Background(), dto.UpdateProfileInput{
			UserID: uuid.New().String(), FirstName: "John", LastName: "Doe", Email: "a@b.com", PhoneNumber: "abc",
		})
		assert.ErrorIs(t, err, domain.ErrInvalidPhoneFormat)
	})

	t.Run("user not found", func(t *testing.T) {
		h := newHarness()
		_, err := h.userSvc.UpdateProfile(context.Background(), dto.UpdateProfileInput{
			UserID: uuid.New().String(), FirstName: "John", LastName: "Doe", Email: "a@b.com",
		})
		assert.ErrorIs(t, err, domain.ErrUserNotFound)
	})

	t.Run("happy path", func(t *testing.T) {
		h := newHarness()
		u := seedUser(h, "bob")
		view, err := h.userSvc.UpdateProfile(context.Background(), dto.UpdateProfileInput{
			UserID: u.ID.String(), FirstName: "Robert", LastName: "Brown", Email: "bob@x.com",
		})
		assert.NoError(t, err)
		assert.Equal(t, "Robert", view.FirstName)
		assert.Equal(t, "Brown", view.LastName)
		assert.Equal(t, "bob@x.com", view.Email)
	})

	t.Run("repo error", func(t *testing.T) {
		h := newHarness()
		u := seedUser(h, "bob")
		h.repo.updateProfileErr = errBoom
		_, err := h.userSvc.UpdateProfile(context.Background(), dto.UpdateProfileInput{
			UserID: u.ID.String(), FirstName: "John", LastName: "Doe", Email: "a@b.com",
		})
		assert.ErrorIs(t, err, errBoom)
	})
}

func TestUserService_ChangePassword(t *testing.T) {
	t.Run("empty new password", func(t *testing.T) {
		h := newHarness()
		err := h.userSvc.ChangePassword(context.Background(), dto.ChangePasswordInput{
			UserID: uuid.New().String(), OldPassword: "old", NewPassword: "  ",
		})
		assert.ErrorIs(t, err, domain.ErrWeakPassword)
	})

	t.Run("invalid UUID", func(t *testing.T) {
		h := newHarness()
		err := h.userSvc.ChangePassword(context.Background(), dto.ChangePasswordInput{
			UserID: "bad", OldPassword: "old", NewPassword: "newpass123",
		})
		assert.ErrorIs(t, err, domain.ErrInvalidInput)
	})

	t.Run("user not found", func(t *testing.T) {
		h := newHarness()
		err := h.userSvc.ChangePassword(context.Background(), dto.ChangePasswordInput{
			UserID: uuid.New().String(), OldPassword: "old", NewPassword: "newpass123",
		})
		assert.ErrorIs(t, err, domain.ErrUserNotFound)
	})

	t.Run("wrong old password", func(t *testing.T) {
		h := newHarness()
		u := seedUser(h, "charlie")
		h.account.verify = false
		err := h.userSvc.ChangePassword(context.Background(), dto.ChangePasswordInput{
			UserID: u.ID.String(), OldPassword: "wrong", NewPassword: "newpass123",
		})
		assert.ErrorIs(t, err, domain.ErrInvalidCredentials)
	})

	t.Run("happy path", func(t *testing.T) {
		h := newHarness()
		u := seedUser(h, "dave")
		h.account.verify = true
		err := h.userSvc.ChangePassword(context.Background(), dto.ChangePasswordInput{
			UserID: u.ID.String(), OldPassword: "old", NewPassword: "newpass123",
		})
		assert.NoError(t, err)
	})

	t.Run("hash error", func(t *testing.T) {
		h := newHarness()
		u := seedUser(h, "dave")
		h.account.verify = true
		h.account.hashErr = errBoom
		err := h.userSvc.ChangePassword(context.Background(), dto.ChangePasswordInput{
			UserID: u.ID.String(), OldPassword: "old", NewPassword: "newpass123",
		})
		assert.ErrorIs(t, err, errBoom)
	})

	t.Run("update error", func(t *testing.T) {
		h := newHarness()
		u := seedUser(h, "dave")
		h.account.verify = true
		h.repo.updatePassErr = errBoom
		err := h.userSvc.ChangePassword(context.Background(), dto.ChangePasswordInput{
			UserID: u.ID.String(), OldPassword: "old", NewPassword: "newpass123",
		})
		assert.ErrorIs(t, err, errBoom)
	})
}

func TestUserService_GetBalance(t *testing.T) {
	t.Run("invalid UUID", func(t *testing.T) {
		h := newHarness()
		_, err := h.userSvc.GetBalance(context.Background(), "bad")
		assert.ErrorIs(t, err, domain.ErrInvalidInput)
	})

	t.Run("happy path", func(t *testing.T) {
		h := newHarness()
		uid := uuid.New()
		h.balance.balances[uid] = &model.PointsBalance{UserID: uid, Points: 500, UpdatedAt: time.Now()}
		view, err := h.userSvc.GetBalance(context.Background(), uid.String())
		assert.NoError(t, err)
		assert.Equal(t, int64(500), view.Points)
	})
}

func TestUserService_GetTransactions(t *testing.T) {
	t.Run("invalid UUID", func(t *testing.T) {
		h := newHarness()
		_, _, err := h.userSvc.GetTransactions(context.Background(), dto.GetTransactionsInput{UserID: "bad"})
		assert.ErrorIs(t, err, domain.ErrInvalidInput)
	})

	t.Run("happy path", func(t *testing.T) {
		h := newHarness()
		uid := uuid.New()
		txID := uuid.New()
		opID := uuid.New()
		h.balance.txs = []model.PointsTransaction{
			{ID: txID, UserID: uid, OperationID: opID, Amount: 100, Reason: "test", CreatedAt: time.Now()},
		}
		views, token, err := h.userSvc.GetTransactions(context.Background(), dto.GetTransactionsInput{
			UserID: uid.String(), PageSize: 10,
		})
		assert.NoError(t, err)
		assert.Len(t, views, 1)
		assert.Equal(t, txID.String(), views[0].ID)
		assert.Equal(t, "", token)
	})

	t.Run("repo error", func(t *testing.T) {
		h := newHarness()
		h.balance.getTransErr = errBoom
		_, _, err := h.userSvc.GetTransactions(context.Background(), dto.GetTransactionsInput{
			UserID: uuid.New().String(),
		})
		assert.ErrorIs(t, err, errBoom)
	})
}

func TestUserService_DeductPoints(t *testing.T) {
	t.Run("invalid user UUID", func(t *testing.T) {
		h := newHarness()
		_, err := h.userSvc.DeductPoints(context.Background(), dto.DeductPointsInput{
			UserID: "bad", OperationID: uuid.New().String(), Amount: 100, Reason: "test",
		})
		assert.ErrorIs(t, err, domain.ErrInvalidInput)
	})

	t.Run("invalid operation UUID", func(t *testing.T) {
		h := newHarness()
		_, err := h.userSvc.DeductPoints(context.Background(), dto.DeductPointsInput{
			UserID: uuid.New().String(), OperationID: "bad", Amount: 100, Reason: "test",
		})
		assert.ErrorIs(t, err, domain.ErrInvalidInput)
	})

	t.Run("happy path", func(t *testing.T) {
		h := newHarness()
		uid := uuid.New()
		h.balance.balances[uid] = &model.PointsBalance{UserID: uid, Points: 500, UpdatedAt: time.Now()}
		view, err := h.userSvc.DeductPoints(context.Background(), dto.DeductPointsInput{
			UserID: uid.String(), OperationID: uuid.New().String(), Amount: 100, Reason: "test",
		})
		assert.NoError(t, err)
		assert.Equal(t, int64(400), view.Points)
	})

	t.Run("insufficient balance", func(t *testing.T) {
		h := newHarness()
		uid := uuid.New()
		h.balance.balances[uid] = &model.PointsBalance{UserID: uid, Points: 50, UpdatedAt: time.Now()}
		_, err := h.userSvc.DeductPoints(context.Background(), dto.DeductPointsInput{
			UserID: uid.String(), OperationID: uuid.New().String(), Amount: 100, Reason: "test",
		})
		assert.ErrorIs(t, err, domain.ErrInsufficientBalance)
	})

	t.Run("repo error", func(t *testing.T) {
		h := newHarness()
		h.balance.applyErr = errBoom
		_, err := h.userSvc.DeductPoints(context.Background(), dto.DeductPointsInput{
			UserID: uuid.New().String(), OperationID: uuid.New().String(), Amount: 100, Reason: "test",
		})
		assert.ErrorIs(t, err, errBoom)
	})

	t.Run("with order ID", func(t *testing.T) {
		h := newHarness()
		uid := uuid.New()
		oid := uuid.New()
		h.balance.balances[uid] = &model.PointsBalance{UserID: uid, Points: 500, UpdatedAt: time.Now()}
		view, err := h.userSvc.DeductPoints(context.Background(), dto.DeductPointsInput{
			UserID: uid.String(), OperationID: uuid.New().String(), OrderID: oid.String(), Amount: 200, Reason: "order",
		})
		assert.NoError(t, err)
		assert.Equal(t, int64(300), view.Points)
	})
}

func TestUserService_AddPoints(t *testing.T) {
	t.Run("invalid UUID", func(t *testing.T) {
		h := newHarness()
		_, err := h.userSvc.AddPoints(context.Background(), dto.AddPointsInput{
			UserID: "bad", OperationID: uuid.New().String(), Amount: 100, Reason: "test",
		})
		assert.ErrorIs(t, err, domain.ErrInvalidInput)
	})

	t.Run("happy path", func(t *testing.T) {
		h := newHarness()
		uid := uuid.New()
		h.balance.balances[uid] = &model.PointsBalance{UserID: uid, Points: 100, UpdatedAt: time.Now()}
		view, err := h.userSvc.AddPoints(context.Background(), dto.AddPointsInput{
			UserID: uid.String(), OperationID: uuid.New().String(), Amount: 50, Reason: "bonus",
		})
		assert.NoError(t, err)
		assert.Equal(t, int64(150), view.Points)
	})

	t.Run("repo error", func(t *testing.T) {
		h := newHarness()
		h.balance.applyErr = errBoom
		_, err := h.userSvc.AddPoints(context.Background(), dto.AddPointsInput{
			UserID: uuid.New().String(), OperationID: uuid.New().String(), Amount: 50, Reason: "bonus",
		})
		assert.ErrorIs(t, err, errBoom)
	})
}
