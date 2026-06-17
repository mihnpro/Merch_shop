package e2e

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestAuthRegister(t *testing.T) {
	c := newClient(cfg)
	login := uniqueLogin("e2e_reg")
	email := login + "@example.com"

	t.Run("happy path", func(t *testing.T) {
		status, raw, err := c.do(http.MethodPost, "/auth/register", registerBody{
			Login:       login,
			Password:    seedPassword,
			FirstName:   "Test",
			LastName:    "User",
			Email:       email,
			PhoneNumber: "+1234567890",
			Patronymic:  "Patronymic",
		})
		require.NoError(t, err)
		requireStatus(t, http.StatusCreated, status, raw)
		u := decode[userView](t, raw)
		trackUser(u.ID)
		require.NotEmpty(t, u.ID)
		require.Equal(t, login, u.Login)
		require.Equal(t, "user", u.Role)
		require.Equal(t, "active", u.Status)
	})

	t.Run("duplicate login → 409", func(t *testing.T) {
		status, raw, err := c.do(http.MethodPost, "/auth/register", registerBody{
			Login: login, Password: seedPassword,
			FirstName: "X", LastName: "Y", Email: "dup@example.com",
		})
		require.NoError(t, err)
		requireStatus(t, http.StatusConflict, status, raw)
	})

	t.Run("invalid email → 400", func(t *testing.T) {
		status, raw, err := c.do(http.MethodPost, "/auth/register", registerBody{
			Login: uniqueLogin("e2e_bademail"), Password: seedPassword,
			FirstName: "A", LastName: "B", Email: "not-an-email",
		})
		require.NoError(t, err)
		requireStatus(t, http.StatusBadRequest, status, raw)
	})

	t.Run("weak password → 400", func(t *testing.T) {
		status, raw, err := c.do(http.MethodPost, "/auth/register", registerBody{
			Login: uniqueLogin("e2e_weakpw"), Password: "123",
			FirstName: "A", LastName: "B", Email: "w@x.com",
		})
		require.NoError(t, err)
		requireStatus(t, http.StatusBadRequest, status, raw)
	})

	t.Run("empty first name → 400", func(t *testing.T) {
		status, raw, err := c.do(http.MethodPost, "/auth/register", registerBody{
			Login: uniqueLogin("e2e_noname"), Password: seedPassword,
			FirstName: "", LastName: "B", Email: "n@x.com",
		})
		require.NoError(t, err)
		requireStatus(t, http.StatusBadRequest, status, raw)
	})

	t.Run("broken JSON → 400", func(t *testing.T) {
		status, raw, err := c.doRaw(http.MethodPost, "/auth/register", "{")
		require.NoError(t, err)
		requireStatus(t, http.StatusBadRequest, status, raw)
	})
}

func TestAuthLogin(t *testing.T) {
	c, u := newUserClient(t)

	t.Run("happy path returns user", func(t *testing.T) {
		status, raw, err := newClient(cfg).do(http.MethodPost, "/auth/login", loginBody{
			Login: u.Login, Password: seedPassword,
		})
		require.NoError(t, err)
		requireStatus(t, http.StatusOK, status, raw)
		v := decode[loginView](t, raw)
		require.Equal(t, u.Login, v.User.Login)
	})

	t.Run("captured token authenticates", func(t *testing.T) {
		require.NotEmpty(t, c.token, "login should capture the access token")
	})

	t.Run("wrong password → 401", func(t *testing.T) {
		status, raw, err := newClient(cfg).do(http.MethodPost, "/auth/login", loginBody{
			Login: u.Login, Password: "wrongpass",
		})
		require.NoError(t, err)
		requireStatus(t, http.StatusUnauthorized, status, raw)
	})

	t.Run("unknown user → 401", func(t *testing.T) {
		status, raw, err := newClient(cfg).do(http.MethodPost, "/auth/login", loginBody{
			Login: "ghost_user_99999", Password: seedPassword,
		})
		require.NoError(t, err)
		requireStatus(t, http.StatusUnauthorized, status, raw)
	})

	t.Run("empty credentials → 401", func(t *testing.T) {
		status, raw, err := newClient(cfg).do(http.MethodPost, "/auth/login", loginBody{
			Login: "  ", Password: "",
		})
		require.NoError(t, err)
		requireStatus(t, http.StatusUnauthorized, status, raw)
	})
}

func TestAuthRefreshAndLogout(t *testing.T) {
	t.Run("refresh without cookie → 401", func(t *testing.T) {
		status, raw, err := newClient(cfg).do(http.MethodPost, "/auth/refresh", nil)
		require.NoError(t, err)
		requireStatus(t, http.StatusUnauthorized, status, raw)
	})

	t.Run("refresh with cookie → 204", func(t *testing.T) {
		c, _ := newUserClient(t)
		status, raw, err := c.do(http.MethodPost, "/auth/refresh", nil)
		require.NoError(t, err)
		requireStatus(t, http.StatusNoContent, status, raw)
	})

	t.Run("logout → 204", func(t *testing.T) {
		c, _ := newUserClient(t)
		status, raw, err := c.do(http.MethodPost, "/auth/logout", nil)
		require.NoError(t, err)
		requireStatus(t, http.StatusNoContent, status, raw)
	})
}
