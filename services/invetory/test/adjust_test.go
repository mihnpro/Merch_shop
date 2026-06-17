package e2e

import (
	"net/http"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func TestAdjustStockHappyPath(t *testing.T) {
	admin := adminClient(t)
	productID := newProductID()
	opID := uuid.NewString()

	t.Run("positive adjustment sets available", func(t *testing.T) {
		status, raw := adjust(t, admin, productID, 50, opID)
		requireStatus(t, http.StatusOK, status, raw)
		v := decode[adjustStockView](t, raw)
		assert.Equal(t, productID, v.ProductID)
		assert.Equal(t, 50, v.Available)
		assert.GreaterOrEqual(t, v.Version, 1)
	})

	t.Run("replaying same operation_id is idempotent", func(t *testing.T) {
		status, raw := adjust(t, admin, productID, 50, opID)
		requireStatus(t, http.StatusOK, status, raw)
		assert.Equal(t, 50, decode[adjustStockView](t, raw).Available, "replay must not double-apply")
	})

	t.Run("new operation accumulates", func(t *testing.T) {
		status, raw := adjust(t, admin, productID, 30, uuid.NewString())
		requireStatus(t, http.StatusOK, status, raw)
		assert.Equal(t, 80, decode[adjustStockView](t, raw).Available)
	})

	t.Run("negative adjustment within available", func(t *testing.T) {
		status, raw := adjust(t, admin, productID, -20, uuid.NewString())
		requireStatus(t, http.StatusOK, status, raw)
		assert.Equal(t, 60, decode[adjustStockView](t, raw).Available)
	})
}

func TestAdjustStockValidation(t *testing.T) {
	admin := adminClient(t)

	t.Run("zero delta → 400", func(t *testing.T) {
		status, raw := adjust(t, admin, newProductID(), 0, uuid.NewString())
		requireStatus(t, http.StatusBadRequest, status, raw)
	})

	t.Run("invalid product id → 400", func(t *testing.T) {
		status, raw := adjust(t, admin, "not-a-uuid", 10, uuid.NewString())
		requireStatus(t, http.StatusBadRequest, status, raw)
	})

	t.Run("invalid operation id → 400", func(t *testing.T) {
		status, raw := adjust(t, admin, newProductID(), 10, "not-a-uuid")
		requireStatus(t, http.StatusBadRequest, status, raw)
	})

	t.Run("empty reason → 400", func(t *testing.T) {
		status, raw, err := admin.do(http.MethodPost, "/admin/inventory/adjust", adjustStockBody{
			ProductID: newProductID(), Delta: 10, OperationID: uuid.NewString(), Reason: "",
		})
		if err != nil {
			t.Fatal(err)
		}
		requireStatus(t, http.StatusBadRequest, status, raw)
	})

	t.Run("decrement below zero → 409", func(t *testing.T) {
		status, raw := adjust(t, admin, newProductID(), -10, uuid.NewString())
		requireStatus(t, http.StatusConflict, status, raw)
	})
}
