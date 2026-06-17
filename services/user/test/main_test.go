package e2e

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"sync/atomic"
	"testing"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/require"
)

var (
	cfg Config
	db  *sqlx.DB

	adminLogin string
	adminPass  string
	adminID    string
)

const seedPassword = "password123"

func TestMain(m *testing.M) {
	cfg = loadConfig()

	if err := waitForReady(cfg); err != nil {
		fmt.Fprintf(os.Stderr, "e2e setup: stack not ready: %v\nhint: bring the stack up first (task up)\n", err)
		os.Exit(1)
	}

	var err error
	db, err = connectDB(cfg)
	if err != nil {
		fmt.Fprintf(os.Stderr, "e2e setup: %v\n", err)
		os.Exit(1)
	}

	if err := seedAdmin(); err != nil {
		fmt.Fprintf(os.Stderr, "e2e setup: seed admin: %v\n", err)
		_ = db.Close()
		os.Exit(1)
	}

	code := m.Run()

	if err := cleanup(db); err != nil {
		fmt.Fprintf(os.Stderr, "e2e teardown: cleanup failed: %v\n", err)
	}
	_ = db.Close()

	os.Exit(code)
}

func waitForReady(cfg Config) error {
	c := newClient(cfg)
	deadline := time.Now().Add(cfg.ReadyTimeout)
	var lastErr error
	for time.Now().Before(deadline) {
		status, _, err := c.login("readiness_probe_nonexistent", "x")
		if err == nil && status == http.StatusUnauthorized {
			return nil
		}
		if err != nil {
			lastErr = err
		} else {
			lastErr = fmt.Errorf("unexpected readiness status %d", status)
		}
		time.Sleep(time.Second)
	}
	return fmt.Errorf("not ready after %s: %w", cfg.ReadyTimeout, lastErr)
}

func seedAdmin() error {
	c := newClient(cfg)
	login := uniqueLogin("e2e_admin")
	status, raw, err := c.do(http.MethodPost, "/auth/register", registerBody{
		Login:     login,
		Password:  seedPassword,
		FirstName: "E2E",
		LastName:  "Admin",
		Email:     login + "@example.com",
	})
	if err != nil {
		return err
	}
	if status != http.StatusCreated {
		return fmt.Errorf("register admin: status %d body %s", status, raw)
	}
	var u userView
	if err := json.Unmarshal(raw, &u); err != nil {
		return fmt.Errorf("decode admin: %w", err)
	}
	trackUser(u.ID)

	if err := promoteToAdmin(db, login); err != nil {
		return err
	}

	adminLogin, adminPass, adminID = login, seedPassword, u.ID
	return nil
}

var loginCounter atomic.Uint64

func uniqueLogin(prefix string) string {
	return fmt.Sprintf("%s_%d_%d", prefix, time.Now().UnixNano(), loginCounter.Add(1))
}

func requireStatus(t *testing.T, want, got int, raw []byte) {
	t.Helper()
	require.Equalf(t, want, got, "unexpected status; body: %s", raw)
}

func decode[T any](t *testing.T, raw []byte) T {
	t.Helper()
	var out T
	require.NoErrorf(t, json.Unmarshal(raw, &out), "decode body: %s", raw)
	return out
}

func newUserClient(t *testing.T) (*Client, userView) {
	t.Helper()
	c := newClient(cfg)
	login := uniqueLogin("e2e_user")
	status, raw, err := c.do(http.MethodPost, "/auth/register", registerBody{
		Login:       login,
		Password:    seedPassword,
		FirstName:   "Test",
		LastName:    "User",
		Email:       login + "@example.com",
		PhoneNumber: "+1234567890",
	})
	require.NoError(t, err)
	requireStatus(t, http.StatusCreated, status, raw)
	u := decode[userView](t, raw)
	trackUser(u.ID)

	status, raw, err = c.login(login, seedPassword)
	require.NoError(t, err)
	requireStatus(t, http.StatusOK, status, raw)
	return c, u
}

func adminClient(t *testing.T) *Client {
	t.Helper()
	c := newClient(cfg)
	status, raw, err := c.login(adminLogin, adminPass)
	require.NoError(t, err)
	requireStatus(t, http.StatusOK, status, raw)
	return c
}
