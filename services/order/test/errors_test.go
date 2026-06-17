package e2e

import (
	"net/http"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

func TestCreateOrderErrors(t *testing.T) {
	t.Run("empty cart → 409", func(t *testing.T) {
		user, _ := newUserClient(t)
		status, raw, err := user.do(http.MethodPost, "/orders", createOrderBody{DeliveryAddress: "123 Main St"})
		require.NoError(t, err)
		requireStatus(t, http.StatusConflict, status, raw)
	})
}

func TestGetOrderErrors(t *testing.T) {
	user, _ := newUserClient(t)

	t.Run("nonexistent order → 404", func(t *testing.T) {
		status, raw, err := user.do(http.MethodGet, "/orders/"+uuid.NewString(), nil)
		require.NoError(t, err)
		requireStatus(t, http.StatusNotFound, status, raw)
	})

	t.Run("malformed order id → 400", func(t *testing.T) {
		status, raw, err := user.do(http.MethodGet, "/orders/not-a-uuid", nil)
		require.NoError(t, err)
		requireStatus(t, http.StatusBadRequest, status, raw)
	})
}

func TestUpdateStatusErrors(t *testing.T) {
	admin := adminClient(t)

	t.Run("nonexistent order → 404", func(t *testing.T) {
		status, raw, err := admin.do(http.MethodPut, "/admin/orders/"+uuid.NewString()+"/status", updateOrderStatusBody{Status: "confirmed"})
		require.NoError(t, err)
		requireStatus(t, http.StatusNotFound, status, raw)
	})
}
