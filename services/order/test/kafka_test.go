package e2e

import (
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

type stockItem struct {
	ProductID string `json:"product_id"`
	Available int    `json:"available"`
}

type stockList struct {
	Items []stockItem `json:"items"`
}

func TestKafkaOrderReservesStock(t *testing.T) {
	admin := adminClient(t)
	user, userID := newUserClient(t)

	const (
		price = 100
		stock = 20
		qty   = 3
	)
	product := createProduct(t, admin, price)
	stockUp(t, admin, product.ID, stock)
	grantPoints(t, admin, userID, price*qty+1000)
	addToCart(t, user, product.ID, qty)

	status, raw, err := user.do(http.MethodPost, "/orders", createOrderBody{DeliveryAddress: "123 Main St"})
	require.NoError(t, err)
	requireStatus(t, http.StatusCreated, status, raw)
	order := decode[orderView](t, raw)
	trackOrder(order.ID)
	require.Equal(t, "pending", order.Status)

	t.Run("inventory reserves stock via kafka event", func(t *testing.T) {
		available, ok := waitForStock(t, admin, product.ID, stock-qty, 40*time.Second)
		require.Truef(t, ok, "stock not reserved within timeout (want available=%d, last=%d); is Kafka + inventory consumer running?", stock-qty, available)
	})

	t.Run("order confirmed by inventory callback", func(t *testing.T) {
		st, ok := waitForOrderStatus(t, user, order.ID, "confirmed", 20*time.Second)
		require.Truef(t, ok, "order not confirmed within timeout (last status=%q)", st)
	})
}

func waitForStock(t *testing.T, c *Client, productID string, want int, timeout time.Duration) (int, bool) {
	t.Helper()
	deadline := time.Now().Add(timeout)
	last := -1
	for time.Now().Before(deadline) {
		status, raw, err := c.do(http.MethodGet, "/stock?product_ids="+productID, nil)
		require.NoError(t, err)
		requireStatus(t, http.StatusOK, status, raw)
		for _, it := range decode[stockList](t, raw).Items {
			if it.ProductID == productID {
				last = it.Available
				if it.Available == want {
					return last, true
				}
			}
		}
		time.Sleep(time.Second)
	}
	return last, false
}

func waitForOrderStatus(t *testing.T, c *Client, orderID, want string, timeout time.Duration) (string, bool) {
	t.Helper()
	deadline := time.Now().Add(timeout)
	last := ""
	for time.Now().Before(deadline) {
		status, raw, err := c.do(http.MethodGet, "/orders/"+orderID, nil)
		require.NoError(t, err)
		requireStatus(t, http.StatusOK, status, raw)
		last = decode[orderView](t, raw).Status
		if last == want {
			return last, true
		}
		time.Sleep(time.Second)
	}
	return last, false
}
