package e2e

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestMeProfile(t *testing.T) {
	c, u := newUserClient(t)

	t.Run("me happy path", func(t *testing.T) {
		status, raw, err := c.do(http.MethodGet, "/me", nil)
		require.NoError(t, err)
		requireStatus(t, http.StatusOK, status, raw)
		me := decode[meView](t, raw)
		require.NotEmpty(t, me.UserID)
		require.Equal(t, "user", me.Role)
	})

	t.Run("me without token → 401", func(t *testing.T) {
		status, raw, err := newClient(cfg).do(http.MethodGet, "/me", nil)
		require.NoError(t, err)
		requireStatus(t, http.StatusUnauthorized, status, raw)
	})

	t.Run("get profile returns login", func(t *testing.T) {
		status, raw, err := c.do(http.MethodGet, "/me/profile", nil)
		require.NoError(t, err)
		requireStatus(t, http.StatusOK, status, raw)
		require.Equal(t, u.Login, decode[userView](t, raw).Login)
	})
}

func TestMeUpdateProfile(t *testing.T) {
	c, u := newUserClient(t)

	t.Run("happy path", func(t *testing.T) {
		status, raw, err := c.do(http.MethodPut, "/me/profile", updateProfileBody{
			FirstName:   "Updated",
			LastName:    "Name",
			Email:       u.Login + "@example.com",
			PhoneNumber: "+1987654321",
		})
		require.NoError(t, err)
		requireStatus(t, http.StatusOK, status, raw)
		require.Equal(t, "Updated", decode[userView](t, raw).FirstName)
	})

	t.Run("empty name → 400", func(t *testing.T) {
		status, raw, err := c.do(http.MethodPut, "/me/profile", updateProfileBody{
			FirstName: "", LastName: "X", Email: "a@b.com",
		})
		require.NoError(t, err)
		requireStatus(t, http.StatusBadRequest, status, raw)
	})

	t.Run("invalid email → 400", func(t *testing.T) {
		status, raw, err := c.do(http.MethodPut, "/me/profile", updateProfileBody{
			FirstName: "A", LastName: "B", Email: "bad",
		})
		require.NoError(t, err)
		requireStatus(t, http.StatusBadRequest, status, raw)
	})
}

func TestMeChangePassword(t *testing.T) {
	c, u := newUserClient(t)
	const newPass = "newpass456"

	t.Run("happy path → 204", func(t *testing.T) {
		status, raw, err := c.do(http.MethodPost, "/me/password", changePasswordBody{
			OldPassword: seedPassword, NewPassword: newPass,
		})
		require.NoError(t, err)
		requireStatus(t, http.StatusNoContent, status, raw)
	})

	t.Run("wrong old → 401", func(t *testing.T) {
		status, raw, err := c.do(http.MethodPost, "/me/password", changePasswordBody{
			OldPassword: "wrongold", NewPassword: "newpass789",
		})
		require.NoError(t, err)
		requireStatus(t, http.StatusUnauthorized, status, raw)
	})

	t.Run("weak new → 400", func(t *testing.T) {
		status, raw, err := c.do(http.MethodPost, "/me/password", changePasswordBody{
			OldPassword: newPass, NewPassword: "123",
		})
		require.NoError(t, err)
		requireStatus(t, http.StatusBadRequest, status, raw)
	})

	t.Run("re-login with new password", func(t *testing.T) {
		status, raw, err := newClient(cfg).login(u.Login, newPass)
		require.NoError(t, err)
		requireStatus(t, http.StatusOK, status, raw)
	})
}

func TestMeBalanceAndTransactions(t *testing.T) {
	c, _ := newUserClient(t)

	t.Run("balance starts at zero", func(t *testing.T) {
		status, raw, err := c.do(http.MethodGet, "/me/balance", nil)
		require.NoError(t, err)
		requireStatus(t, http.StatusOK, status, raw)
		require.EqualValues(t, 0, decode[balanceView](t, raw).Points)
	})

	t.Run("transactions is an array", func(t *testing.T) {
		status, raw, err := c.do(http.MethodGet, "/me/transactions", nil)
		require.NoError(t, err)
		requireStatus(t, http.StatusOK, status, raw)
		require.NotNil(t, decode[transactionsView](t, raw).Transactions)
	})
}
