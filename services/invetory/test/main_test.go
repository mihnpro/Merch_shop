package e2e

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"sync/atomic"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/require"
)

var (
	cfg         Config
	userDB      *sqlx.DB
	inventoryDB *sqlx.DB

	adminLogin string
	adminPass  string
)

const seedPassword = "password123"

func TestMain(m *testing.M) {
	cfg = loadConfig()

	if err := waitForReady(cfg); err != nil {
		fmt.Fprintf(os.Stderr, "e2e setup: stack not ready: %v\nhint: bring the stack up first (task up)\n", err)
		os.Exit(1)
	}

	var err error
	if userDB, err = connect(cfg.UserDB); err != nil {
		fmt.Fprintf(os.Stderr, "e2e setup: %v\n", err)
		os.Exit(1)
	}
	if inventoryDB, err = connect(cfg.InventoryDB); err != nil {
		fmt.Fprintf(os.Stderr, "e2e setup: %v\n", err)
		_ = userDB.Close()
		os.Exit(1)
	}

	if err := seedAdmin(); err != nil {
		fmt.Fprintf(os.Stderr, "e2e setup: seed admin: %v\n", err)
		_ = userDB.Close()
		_ = inventoryDB.Close()
		os.Exit(1)
	}

	code := m.Run()

	if err := cleanup(userDB, inventoryDB); err != nil {
		fmt.Fprintf(os.Stderr, "e2e teardown: cleanup failed: %v\n", err)
	}
	_ = userDB.Close()
	_ = inventoryDB.Close()

	os.Exit(code)
}

func waitForReady(cfg Config) error {
	c := newClient(cfg)
	deadline := time.Now().Add(cfg.ReadyTimeout)
	var lastErr error
	for time.Now().Before(deadline) {
		authStatus, _, authErr := c.login("readiness_probe_nonexistent", "x")
		stockStatus, _, stockErr := c.do(http.MethodGet, "/stock", nil)
		if authErr == nil && authStatus == http.StatusUnauthorized &&
			stockErr == nil && stockStatus == http.StatusUnauthorized {
			return nil
		}
		switch {
		case authErr != nil:
			lastErr = authErr
		case stockErr != nil:
			lastErr = stockErr
		default:
			lastErr = fmt.Errorf("unexpected readiness statuses auth=%d stock=%d", authStatus, stockStatus)
		}
		time.Sleep(time.Second)
	}
	return fmt.Errorf("not ready after %s: %w", cfg.ReadyTimeout, lastErr)
}

func seedAdmin() error {
	c := newClient(cfg)
	login := uniqueLogin("e2e_iadmin")
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

	if err := promoteToAdmin(userDB, login); err != nil {
		return err
	}
	adminLogin, adminPass = login, seedPassword
	return nil
}

var counter atomic.Uint64

func uniqueLogin(prefix string) string {
	return fmt.Sprintf("%s_%d_%d", prefix, time.Now().UnixNano(), counter.Add(1))
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

func adminClient(t *testing.T) *Client {
	t.Helper()
	c := newClient(cfg)
	status, raw, err := c.login(adminLogin, adminPass)
	require.NoError(t, err)
	requireStatus(t, http.StatusOK, status, raw)
	return c
}

func newUserClient(t *testing.T) *Client {
	t.Helper()
	c := newClient(cfg)
	login := uniqueLogin("e2e_iuser")
	status, raw, err := c.do(http.MethodPost, "/auth/register", registerBody{
		Login:     login,
		Password:  seedPassword,
		FirstName: "Test",
		LastName:  "User",
		Email:     login + "@example.com",
	})
	require.NoError(t, err)
	requireStatus(t, http.StatusCreated, status, raw)
	trackUser(decode[userView](t, raw).ID)

	status, raw, err = c.login(login, seedPassword)
	require.NoError(t, err)
	requireStatus(t, http.StatusOK, status, raw)
	return c
}

func adjust(t *testing.T, admin *Client, productID string, delta int, operationID string) (int, []byte) {
	t.Helper()
	status, raw, err := admin.do(http.MethodPost, "/admin/inventory/adjust", adjustStockBody{
		ProductID:   productID,
		Delta:       delta,
		OperationID: operationID,
		Reason:      "e2e adjustment",
	})
	require.NoError(t, err)
	return status, raw
}

func newProductID() string {
	id := uuid.NewString()
	trackProduct(id)
	return id
}
