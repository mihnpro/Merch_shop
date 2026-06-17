package e2e

import (
	"net/http"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

const zeroUUID = "00000000-0000-0000-0000-000000000000"

func TestAdminListUsers(t *testing.T) {
	admin := adminClient(t)
	_, target := newUserClient(t)

	t.Run("list users → array", func(t *testing.T) {
		status, raw, err := admin.do(http.MethodGet, "/admin/users", nil)
		require.NoError(t, err)
		requireStatus(t, http.StatusOK, status, raw)
		require.NotNil(t, decode[listUsersView](t, raw).Users)
	})

	t.Run("search finds our user", func(t *testing.T) {
		status, raw, err := admin.do(http.MethodGet, "/admin/users?search="+target.Login, nil)
		require.NoError(t, err)
		requireStatus(t, http.StatusOK, status, raw)
		users := decode[listUsersView](t, raw).Users
		require.Len(t, users, 1)
		require.Equal(t, target.Login, users[0].Login)
	})

	t.Run("filter by role", func(t *testing.T) {
		status, raw, err := admin.do(http.MethodGet, "/admin/users?role=user", nil)
		require.NoError(t, err)
		requireStatus(t, http.StatusOK, status, raw)
	})
}

func TestAdminGetUser(t *testing.T) {
	admin := adminClient(t)
	_, target := newUserClient(t)

	t.Run("by id → 200", func(t *testing.T) {
		status, raw, err := admin.do(http.MethodGet, "/admin/users/"+target.ID, nil)
		require.NoError(t, err)
		requireStatus(t, http.StatusOK, status, raw)
		require.Equal(t, target.ID, decode[userView](t, raw).ID)
	})

	t.Run("not found → 404", func(t *testing.T) {
		status, raw, err := admin.do(http.MethodGet, "/admin/users/"+zeroUUID, nil)
		require.NoError(t, err)
		requireStatus(t, http.StatusNotFound, status, raw)
	})
}

func TestAdminGrantPoints(t *testing.T) {
	admin := adminClient(t)
	user, target := newUserClient(t)
	opID := uuid.NewString()

	t.Run("grant → 200", func(t *testing.T) {
		status, raw, err := admin.do(http.MethodPost, "/admin/users/"+target.ID+"/grant-points", grantPointsBody{
			Amount: 500, OperationID: opID, Reason: "test bonus",
		})
		require.NoError(t, err)
		requireStatus(t, http.StatusOK, status, raw)
		require.EqualValues(t, 500, decode[balanceView](t, raw).Points)
	})

	t.Run("idempotent on same operation_id", func(t *testing.T) {
		status, raw, err := admin.do(http.MethodPost, "/admin/users/"+target.ID+"/grant-points", grantPointsBody{
			Amount: 500, OperationID: opID, Reason: "test bonus",
		})
		require.NoError(t, err)
		requireStatus(t, http.StatusOK, status, raw)
		require.EqualValues(t, 500, decode[balanceView](t, raw).Points, "replay must not double-credit")
	})

	t.Run("grant to nonexistent → 404", func(t *testing.T) {
		status, raw, err := admin.do(http.MethodPost, "/admin/users/"+zeroUUID+"/grant-points", grantPointsBody{
			Amount: 100, OperationID: uuid.NewString(), Reason: "nope",
		})
		require.NoError(t, err)
		requireStatus(t, http.StatusNotFound, status, raw)
	})

	t.Run("user sees updated balance", func(t *testing.T) {
		status, raw, err := user.do(http.MethodGet, "/me/balance", nil)
		require.NoError(t, err)
		requireStatus(t, http.StatusOK, status, raw)
		require.EqualValues(t, 500, decode[balanceView](t, raw).Points)
	})
}

func TestAdminResetPassword(t *testing.T) {
	admin := adminClient(t)
	_, target := newUserClient(t)

	status, raw, err := admin.do(http.MethodPost, "/admin/users/"+target.ID+"/reset-password", nil)
	require.NoError(t, err)
	requireStatus(t, http.StatusOK, status, raw)
	newPass := decode[resetPasswordView](t, raw).NewPassword
	require.Len(t, newPass, 12)

	status, raw, err = newClient(cfg).login(target.Login, newPass)
	require.NoError(t, err)
	requireStatus(t, http.StatusOK, status, raw)
}

func TestAdminBlockUser(t *testing.T) {
	admin := adminClient(t)
	user, target := newUserClient(t)

	t.Run("block → 200 blocked", func(t *testing.T) {
		status, raw, err := admin.do(http.MethodPut, "/admin/users/"+target.ID+"/status", blockUserBody{Blocked: true})
		require.NoError(t, err)
		requireStatus(t, http.StatusOK, status, raw)
		require.Equal(t, "blocked", decode[userView](t, raw).Status)
	})

	t.Run("blocked user /me → 403", func(t *testing.T) {
		status, raw, err := user.do(http.MethodGet, "/me", nil)
		require.NoError(t, err)
		requireStatus(t, http.StatusForbidden, status, raw)
	})

	t.Run("admin block self → 403", func(t *testing.T) {
		status, raw, err := admin.do(http.MethodPut, "/admin/users/"+adminID+"/status", blockUserBody{Blocked: true})
		require.NoError(t, err)
		requireStatus(t, http.StatusForbidden, status, raw)
	})

	t.Run("unblock → 200 active", func(t *testing.T) {
		status, raw, err := admin.do(http.MethodPut, "/admin/users/"+target.ID+"/status", blockUserBody{Blocked: false})
		require.NoError(t, err)
		requireStatus(t, http.StatusOK, status, raw)
		require.Equal(t, "active", decode[userView](t, raw).Status)
	})
}

func TestAdminChangeRole(t *testing.T) {
	admin := adminClient(t)
	_, target := newUserClient(t)

	t.Run("admin change own role → 403", func(t *testing.T) {
		status, raw, err := admin.do(http.MethodPut, "/admin/users/"+adminID+"/role", changeRoleBody{Role: "user"})
		require.NoError(t, err)
		requireStatus(t, http.StatusForbidden, status, raw)
	})

	t.Run("change role to admin → 200", func(t *testing.T) {
		status, raw, err := admin.do(http.MethodPut, "/admin/users/"+target.ID+"/role", changeRoleBody{Role: "admin"})
		require.NoError(t, err)
		requireStatus(t, http.StatusOK, status, raw)
		require.Equal(t, "admin", decode[userView](t, raw).Role)
	})

	t.Run("change role back to user → 200", func(t *testing.T) {
		status, raw, err := admin.do(http.MethodPut, "/admin/users/"+target.ID+"/role", changeRoleBody{Role: "user"})
		require.NoError(t, err)
		requireStatus(t, http.StatusOK, status, raw)
		require.Equal(t, "user", decode[userView](t, raw).Role)
	})

	t.Run("invalid role → 400", func(t *testing.T) {
		status, raw, err := admin.do(http.MethodPut, "/admin/users/"+target.ID+"/role", changeRoleBody{Role: "superadmin"})
		require.NoError(t, err)
		requireStatus(t, http.StatusBadRequest, status, raw)
	})
}

func TestAdminStatsAndTransactions(t *testing.T) {
	admin := adminClient(t)
	_, target := newUserClient(t)

	for _, period := range []string{"day", "week", "month"} {
		t.Run("stats "+period, func(t *testing.T) {
			status, raw, err := admin.do(http.MethodGet, "/admin/users/stats?period="+period, nil)
			require.NoError(t, err)
			requireStatus(t, http.StatusOK, status, raw)
			require.GreaterOrEqual(t, decode[statsView](t, raw).NewUsers, 0)
		})
	}

	t.Run("user transactions → array", func(t *testing.T) {
		status, raw, err := admin.do(http.MethodGet, "/admin/users/"+target.ID+"/transactions", nil)
		require.NoError(t, err)
		requireStatus(t, http.StatusOK, status, raw)
		require.NotNil(t, decode[transactionsView](t, raw).Transactions)
	})
}
