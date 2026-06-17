package e2e

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCreateOrderHappyPath(t *testing.T) {
	admin := adminClient(t)
	user, userID := newUserClient(t)

	order := placeOrder(t, admin, user, userID, 100, 2)

	assert.Equal(t, "pending", order.Status)
	assert.Equal(t, int64(200), order.TotalPoints)
	assert.Equal(t, userID, order.UserID)
	require.Len(t, order.Items, 1)
	assert.Equal(t, 2, order.Items[0].Quantity)

	t.Run("user can fetch own order", func(t *testing.T) {
		status, raw, err := user.do(http.MethodGet, "/orders/"+order.ID, nil)
		require.NoError(t, err)
		requireStatus(t, http.StatusOK, status, raw)
		assert.Equal(t, order.ID, decode[orderView](t, raw).ID)
	})

	t.Run("order appears in user's list", func(t *testing.T) {
		status, raw, err := user.do(http.MethodGet, "/orders", nil)
		require.NoError(t, err)
		requireStatus(t, http.StatusOK, status, raw)
		list := decode[ordersListView](t, raw)
		found := false
		for _, o := range list.Orders {
			if o.ID == order.ID {
				found = true
			}
		}
		assert.True(t, found, "created order not in user's order list")
	})

	t.Run("cart is cleared after order", func(t *testing.T) {
		status, raw, err := user.do(http.MethodPost, "/orders", createOrderBody{DeliveryAddress: "456 Side St"})
		require.NoError(t, err)
		requireStatus(t, http.StatusConflict, status, raw)
	})
}

func TestOrderStatusTransitions(t *testing.T) {
	admin := adminClient(t)
	user, userID := newUserClient(t)
	order := placeOrder(t, admin, user, userID, 50, 1)

	t.Run("pending → confirmed", func(t *testing.T) {
		status, raw, err := admin.do(http.MethodPut, "/admin/orders/"+order.ID+"/status", updateOrderStatusBody{Status: "confirmed", Reason: "ok"})
		require.NoError(t, err)
		requireStatus(t, http.StatusOK, status, raw)
		assert.Equal(t, "confirmed", decode[orderView](t, raw).Status)
	})

	t.Run("confirmed → ready_to_pickup", func(t *testing.T) {
		status, raw, err := admin.do(http.MethodPut, "/admin/orders/"+order.ID+"/status", updateOrderStatusBody{Status: "ready_to_pickup"})
		require.NoError(t, err)
		requireStatus(t, http.StatusOK, status, raw)
		assert.Equal(t, "ready_to_pickup", decode[orderView](t, raw).Status)
	})

	t.Run("ready_to_pickup → delivered", func(t *testing.T) {
		status, raw, err := admin.do(http.MethodPut, "/admin/orders/"+order.ID+"/status", updateOrderStatusBody{Status: "delivered"})
		require.NoError(t, err)
		requireStatus(t, http.StatusOK, status, raw)
		assert.Equal(t, "delivered", decode[orderView](t, raw).Status)
	})

	t.Run("delivered → cancelled is rejected", func(t *testing.T) {
		status, raw, err := admin.do(http.MethodPut, "/admin/orders/"+order.ID+"/status", updateOrderStatusBody{Status: "cancelled"})
		require.NoError(t, err)
		requireStatus(t, http.StatusConflict, status, raw)
	})
}

func TestCancelOrderRefunds(t *testing.T) {
	admin := adminClient(t)
	user, userID := newUserClient(t)
	order := placeOrder(t, admin, user, userID, 70, 1)

	t.Run("pending → cancelled", func(t *testing.T) {
		status, raw, err := admin.do(http.MethodPut, "/admin/orders/"+order.ID+"/status", updateOrderStatusBody{Status: "cancelled", Reason: "changed mind"})
		require.NoError(t, err)
		requireStatus(t, http.StatusOK, status, raw)
		assert.Equal(t, "cancelled", decode[orderView](t, raw).Status)
	})

	t.Run("points refunded to user", func(t *testing.T) {
		status, raw, err := user.do(http.MethodGet, "/me/balance", nil)
		require.NoError(t, err)
		requireStatus(t, http.StatusOK, status, raw)
		bal := decode[struct {
			Points int64 `json:"points"`
		}](t, raw)
		assert.GreaterOrEqual(t, bal.Points, int64(70), "refund should restore points")
	})
}

func TestAdminListAndAnalytics(t *testing.T) {
	admin := adminClient(t)
	user, userID := newUserClient(t)
	order := placeOrder(t, admin, user, userID, 100, 1)

	t.Run("admin sees order in global list", func(t *testing.T) {
		status, raw, err := admin.do(http.MethodGet, "/admin/orders?user_id="+userID, nil)
		require.NoError(t, err)
		requireStatus(t, http.StatusOK, status, raw)
		list := decode[ordersListView](t, raw)
		found := false
		for _, o := range list.Orders {
			if o.ID == order.ID {
				found = true
			}
		}
		assert.True(t, found, "order missing from admin listing")
	})

	t.Run("analytics responds for each period", func(t *testing.T) {
		for _, period := range []string{"day", "week", "month"} {
			status, raw, err := admin.do(http.MethodGet, "/admin/analytics?period="+period, nil)
			require.NoError(t, err)
			requireStatus(t, http.StatusOK, status, raw)
			assert.Equal(t, period, decode[analyticsView](t, raw).Period)
		}
	})
}
