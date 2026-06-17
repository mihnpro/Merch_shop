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
	cfg Config
	dbs = map[string]*sqlx.DB{}

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

	conns := map[string]pgConfig{
		"user": cfg.UserDB, "products": cfg.ProductsDB, "inventory": cfg.InventoryDB, "cart": cfg.CartDB,
	}
	for name, pc := range conns {
		db, err := connect(pc)
		if err != nil {
			fmt.Fprintf(os.Stderr, "e2e setup: %v\n", err)
			closeAll()
			os.Exit(1)
		}
		dbs[name] = db
	}

	if err := seedAdmin(); err != nil {
		fmt.Fprintf(os.Stderr, "e2e setup: seed admin: %v\n", err)
		closeAll()
		os.Exit(1)
	}

	code := m.Run()

	if err := cleanup(dbs); err != nil {
		fmt.Fprintf(os.Stderr, "e2e teardown: cleanup failed: %v\n", err)
	}
	closeAll()

	os.Exit(code)
}

func closeAll() {
	for _, db := range dbs {
		if db != nil {
			_ = db.Close()
		}
	}
}

func waitForReady(cfg Config) error {
	c := newClient(cfg)
	deadline := time.Now().Add(cfg.ReadyTimeout)
	var lastErr error
	for time.Now().Before(deadline) {
		authStatus, _, authErr := c.login("readiness_probe_nonexistent", "x")
		cartStatus, _, cartErr := c.do(http.MethodGet, "/cart", nil)
		if authErr == nil && authStatus == http.StatusUnauthorized &&
			cartErr == nil && cartStatus == http.StatusUnauthorized {
			return nil
		}
		switch {
		case authErr != nil:
			lastErr = authErr
		case cartErr != nil:
			lastErr = cartErr
		default:
			lastErr = fmt.Errorf("unexpected readiness statuses auth=%d cart=%d", authStatus, cartStatus)
		}
		time.Sleep(time.Second)
	}
	return fmt.Errorf("not ready after %s: %w", cfg.ReadyTimeout, lastErr)
}

func seedAdmin() error {
	c := newClient(cfg)
	login := uniqueLogin("e2e_cadmin")
	status, raw, err := c.do(http.MethodPost, "/auth/register", registerBody{
		Login: login, Password: seedPassword, FirstName: "E2E", LastName: "Admin", Email: login + "@example.com",
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

	if err := promoteToAdmin(dbs["user"], login); err != nil {
		return err
	}
	adminLogin, adminPass = login, seedPassword
	return nil
}

var counter atomic.Uint64

func uniqueLogin(prefix string) string {
	return fmt.Sprintf("%s_%d_%d", prefix, time.Now().UnixNano(), counter.Add(1))
}

func uniqueCode() string {
	n := counter.Add(1) + uint64(time.Now().UnixNano())
	letters := make([]byte, 12)
	for i := range letters {
		letters[i] = byte('a' + n%26)
		n /= 26
	}
	return "eee_" + string(letters)
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
	login := uniqueLogin("e2e_cuser")
	status, raw, err := c.do(http.MethodPost, "/auth/register", registerBody{
		Login: login, Password: seedPassword, FirstName: "Test", LastName: "User", Email: login + "@example.com",
	})
	require.NoError(t, err)
	requireStatus(t, http.StatusCreated, status, raw)
	trackUser(decode[userView](t, raw).ID)

	status, raw, err = c.login(login, seedPassword)
	require.NoError(t, err)
	requireStatus(t, http.StatusOK, status, raw)
	return c
}

func createProduct(t *testing.T, admin *Client, price int64) (productView, string) {
	t.Helper()
	status, raw, err := admin.do(http.MethodPost, "/admin/categories", createCategoryBody{Code: uniqueCode(), Name: "E2E Cat"})
	require.NoError(t, err)
	requireStatus(t, http.StatusCreated, status, raw)
	cat := decode[categoryView](t, raw)
	trackCategory(cat.ID)

	status, raw, err = admin.do(http.MethodPost, "/admin/products", createProductBody{
		Name: "E2E Product", Description: "e2e", PricePoints: price, CategoryID: cat.ID,
	})
	require.NoError(t, err)
	requireStatus(t, http.StatusCreated, status, raw)
	p := decode[productView](t, raw)
	trackProduct(p.ID)
	p.PricePoints = price
	return p, cat.ID
}

func stockUp(t *testing.T, admin *Client, productID string, qty int) {
	t.Helper()
	status, raw, err := admin.do(http.MethodPost, "/admin/inventory/adjust", adjustStockBody{
		ProductID: productID, Delta: qty, OperationID: uuid.NewString(), Reason: "e2e stock",
	})
	require.NoError(t, err)
	requireStatus(t, http.StatusOK, status, raw)
}

func deactivateProduct(t *testing.T, admin *Client, p productView, categoryID string) {
	t.Helper()
	status, raw, err := admin.do(http.MethodPut, "/admin/products/"+p.ID, updateProductBody{
		Name: "E2E Product", Description: "e2e", PricePoints: p.PricePoints, CategoryID: categoryID, Active: false, Version: p.Version,
	})
	require.NoError(t, err)
	requireStatus(t, http.StatusOK, status, raw)
}

func addToCart(t *testing.T, user *Client, productID string, qty int) (int, []byte) {
	t.Helper()
	status, raw, err := user.do(http.MethodPost, "/cart/items", addCartItemBody{ProductID: productID, Quantity: qty})
	require.NoError(t, err)
	return status, raw
}
