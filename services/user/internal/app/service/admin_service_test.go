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
	vo "github.com/mihnpro/Merch_shop/services/user_customer/internal/domain/valueobject"
)

func TestAdminService_GrantPoints(t *testing.T) {
	t.Run("invalid user UUID", func(t *testing.T) {
		h := newHarness()
		_, err := h.adminSvc.GrantPoints(context.Background(), dto.GrantPointsInput{
			UserID: "bad", OperationID: uuid.New().String(), Amount: 100, Reason: "test",
		})
		assert.ErrorIs(t, err, domain.ErrInvalidInput)
	})

	t.Run("user not found", func(t *testing.T) {
		h := newHarness()
		_, err := h.adminSvc.GrantPoints(context.Background(), dto.GrantPointsInput{
			UserID: uuid.New().String(), OperationID: uuid.New().String(), Amount: 100, Reason: "test",
		})
		assert.ErrorIs(t, err, domain.ErrUserNotFound)
	})

	t.Run("invalid operation UUID", func(t *testing.T) {
		h := newHarness()
		u := seedUser(h, "alice")
		_, err := h.adminSvc.GrantPoints(context.Background(), dto.GrantPointsInput{
			UserID: u.ID.String(), OperationID: "bad", Amount: 100, Reason: "test",
		})
		assert.ErrorIs(t, err, domain.ErrInvalidInput)
	})

	t.Run("happy path", func(t *testing.T) {
		h := newHarness()
		u := seedUser(h, "alice")
		view, err := h.adminSvc.GrantPoints(context.Background(), dto.GrantPointsInput{
			UserID: u.ID.String(), OperationID: uuid.New().String(), Amount: 200, Reason: "bonus",
		})
		assert.NoError(t, err)
		assert.Equal(t, int64(200), view.Points)
	})

	t.Run("credit error", func(t *testing.T) {
		h := newHarness()
		u := seedUser(h, "alice")
		h.balance.applyErr = errBoom
		_, err := h.adminSvc.GrantPoints(context.Background(), dto.GrantPointsInput{
			UserID: u.ID.String(), OperationID: uuid.New().String(), Amount: 100, Reason: "test",
		})
		assert.ErrorIs(t, err, errBoom)
	})
}

func TestAdminService_ResetPassword(t *testing.T) {
	t.Run("invalid UUID", func(t *testing.T) {
		h := newHarness()
		_, err := h.adminSvc.ResetPassword(context.Background(), "bad")
		assert.ErrorIs(t, err, domain.ErrInvalidInput)
	})

	t.Run("user not found", func(t *testing.T) {
		h := newHarness()
		_, err := h.adminSvc.ResetPassword(context.Background(), uuid.New().String())
		assert.ErrorIs(t, err, domain.ErrUserNotFound)
	})

	t.Run("happy path returns 12-char hex", func(t *testing.T) {
		h := newHarness()
		u := seedUser(h, "alice")
		pass, err := h.adminSvc.ResetPassword(context.Background(), u.ID.String())
		assert.NoError(t, err)
		assert.Len(t, pass, 12)
		assert.Regexp(t, `^[0-9a-f]{12}$`, pass)
	})

	t.Run("hash error", func(t *testing.T) {
		h := newHarness()
		u := seedUser(h, "alice")
		h.account.hashErr = errBoom
		_, err := h.adminSvc.ResetPassword(context.Background(), u.ID.String())
		assert.ErrorIs(t, err, errBoom)
	})

	t.Run("update error", func(t *testing.T) {
		h := newHarness()
		u := seedUser(h, "alice")
		h.repo.updatePassErr = errBoom
		_, err := h.adminSvc.ResetPassword(context.Background(), u.ID.String())
		assert.ErrorIs(t, err, errBoom)
	})
}

func TestAdminService_BlockUser(t *testing.T) {
	t.Run("invalid UUID", func(t *testing.T) {
		h := newHarness()
		_, err := h.adminSvc.BlockUser(context.Background(), "bad", true)
		assert.ErrorIs(t, err, domain.ErrInvalidInput)
	})

	t.Run("happy block", func(t *testing.T) {
		h := newHarness()
		u := seedUser(h, "alice")
		view, err := h.adminSvc.BlockUser(context.Background(), u.ID.String(), true)
		assert.NoError(t, err)
		assert.Equal(t, "blocked", view.Status)
		assert.True(t, h.tokens.deleteByUserCalled)
	})

	t.Run("happy unblock", func(t *testing.T) {
		h := newHarness()
		u := seedUser(h, "alice")
		u.Status = vo.StatusBlocked
		view, err := h.adminSvc.BlockUser(context.Background(), u.ID.String(), false)
		assert.NoError(t, err)
		assert.Equal(t, "active", view.Status)
		assert.False(t, h.tokens.deleteByUserCalled)
	})

	t.Run("repo error", func(t *testing.T) {
		h := newHarness()
		u := seedUser(h, "alice")
		h.repo.updateStatusErr = errBoom
		_, err := h.adminSvc.BlockUser(context.Background(), u.ID.String(), true)
		assert.ErrorIs(t, err, errBoom)
	})
}

func TestAdminService_ChangeRole(t *testing.T) {
	t.Run("invalid UUID", func(t *testing.T) {
		h := newHarness()
		_, err := h.adminSvc.ChangeRole(context.Background(), "bad", "admin")
		assert.ErrorIs(t, err, domain.ErrInvalidInput)
	})

	t.Run("invalid role", func(t *testing.T) {
		h := newHarness()
		u := seedUser(h, "alice")
		_, err := h.adminSvc.ChangeRole(context.Background(), u.ID.String(), "superadmin")
		assert.ErrorIs(t, err, domain.ErrInvalidRole)
	})

	t.Run("happy path", func(t *testing.T) {
		h := newHarness()
		u := seedUser(h, "alice")
		view, err := h.adminSvc.ChangeRole(context.Background(), u.ID.String(), "admin")
		assert.NoError(t, err)
		assert.Equal(t, "admin", view.Role)
	})

	t.Run("repo error", func(t *testing.T) {
		h := newHarness()
		u := seedUser(h, "alice")
		h.repo.updateRoleErr = errBoom
		_, err := h.adminSvc.ChangeRole(context.Background(), u.ID.String(), "admin")
		assert.ErrorIs(t, err, errBoom)
	})
}

func TestAdminService_ListUsers(t *testing.T) {
	t.Run("happy path with results", func(t *testing.T) {
		h := newHarness()
		u1 := seedUser(h, "alice")
		u2 := seedUser(h, "bob")
		h.repo.listUsers = []*model.User{u1, u2}
		h.repo.listUsersToken = "next-page"

		views, token, err := h.adminSvc.ListUsers(context.Background(), dto.ListUsersInput{PageSize: 10})
		assert.NoError(t, err)
		assert.Len(t, views, 2)
		assert.Equal(t, "next-page", token)
	})

	t.Run("happy path empty", func(t *testing.T) {
		h := newHarness()
		views, token, err := h.adminSvc.ListUsers(context.Background(), dto.ListUsersInput{})
		assert.NoError(t, err)
		assert.Empty(t, views)
		assert.Equal(t, "", token)
	})

	t.Run("repo error", func(t *testing.T) {
		h := newHarness()
		h.repo.listUsersErr = errBoom
		_, _, err := h.adminSvc.ListUsers(context.Background(), dto.ListUsersInput{})
		assert.ErrorIs(t, err, errBoom)
	})
}

func TestAdminService_GetTransactions(t *testing.T) {
	t.Run("invalid UUID", func(t *testing.T) {
		h := newHarness()
		_, _, err := h.adminSvc.GetTransactions(context.Background(), dto.GetTransactionsInput{UserID: "bad"})
		assert.ErrorIs(t, err, domain.ErrInvalidInput)
	})

	t.Run("happy path", func(t *testing.T) {
		h := newHarness()
		uid := uuid.New()
		txID := uuid.New()
		h.balance.txs = []model.PointsTransaction{
			{ID: txID, UserID: uid, Amount: 50, Reason: "grant", CreatedAt: time.Now()},
		}
		views, _, err := h.adminSvc.GetTransactions(context.Background(), dto.GetTransactionsInput{
			UserID: uid.String(),
		})
		assert.NoError(t, err)
		assert.Len(t, views, 1)
		assert.Equal(t, txID.String(), views[0].ID)
	})

	t.Run("repo error", func(t *testing.T) {
		h := newHarness()
		h.balance.getTransErr = errBoom
		_, _, err := h.adminSvc.GetTransactions(context.Background(), dto.GetTransactionsInput{
			UserID: uuid.New().String(),
		})
		assert.ErrorIs(t, err, errBoom)
	})
}

func TestAdminService_NewUsersStats(t *testing.T) {
	t.Run("day", func(t *testing.T) {
		h := newHarness()
		h.repo.countSince = 5
		stats, err := h.adminSvc.NewUsersStats(context.Background(), "day")
		assert.NoError(t, err)
		assert.Equal(t, 5, stats.NewUsers)
	})

	t.Run("week", func(t *testing.T) {
		h := newHarness()
		h.repo.countSince = 20
		stats, err := h.adminSvc.NewUsersStats(context.Background(), "week")
		assert.NoError(t, err)
		assert.Equal(t, 20, stats.NewUsers)
	})

	t.Run("month", func(t *testing.T) {
		h := newHarness()
		h.repo.countSince = 100
		stats, err := h.adminSvc.NewUsersStats(context.Background(), "month")
		assert.NoError(t, err)
		assert.Equal(t, 100, stats.NewUsers)
	})

	t.Run("repo error", func(t *testing.T) {
		h := newHarness()
		h.repo.countSinceErr = errBoom
		_, err := h.adminSvc.NewUsersStats(context.Background(), "day")
		assert.ErrorIs(t, err, errBoom)
	})
}
