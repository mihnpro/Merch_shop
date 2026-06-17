package service

import (
	"strings"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/mihnpro/Merch_shop/services/invetory/internal/domain"
)

func TestStockFactoryNewAdjustment(t *testing.T) {
	f := NewStockFactory()
	opID := uuid.New()
	prodID := uuid.New()

	t.Run("success", func(t *testing.T) {
		a, err := f.NewAdjustment(AdjustStockInput{OperationID: opID, ProductID: prodID, Delta: 5, Reason: "restock"})
		require.NoError(t, err)
		assert.Equal(t, 5, a.Delta.Int())
		assert.Equal(t, "restock", a.Reason.String())
	})

	t.Run("zero delta", func(t *testing.T) {
		_, err := f.NewAdjustment(AdjustStockInput{OperationID: opID, ProductID: prodID, Delta: 0, Reason: "x"})
		assert.ErrorIs(t, err, domain.ErrZeroDelta)
	})

	t.Run("empty reason", func(t *testing.T) {
		_, err := f.NewAdjustment(AdjustStockInput{OperationID: opID, ProductID: prodID, Delta: 1, Reason: "  "})
		assert.ErrorIs(t, err, domain.ErrEmptyReason)
	})

	t.Run("reason too long", func(t *testing.T) {
		_, err := f.NewAdjustment(AdjustStockInput{OperationID: opID, ProductID: prodID, Delta: 1, Reason: strings.Repeat("a", 501)})
		assert.ErrorIs(t, err, domain.ErrReasonTooLong)
	})

	t.Run("nil operation id", func(t *testing.T) {
		_, err := f.NewAdjustment(AdjustStockInput{OperationID: uuid.Nil, ProductID: prodID, Delta: 1, Reason: "x"})
		assert.ErrorIs(t, err, domain.ErrInvalidOperationID)
	})

	t.Run("nil product id", func(t *testing.T) {
		_, err := f.NewAdjustment(AdjustStockInput{OperationID: opID, ProductID: uuid.Nil, Delta: 1, Reason: "x"})
		assert.ErrorIs(t, err, domain.ErrInvalidProductID)
	})
}

func TestReservationFactoryCreate(t *testing.T) {
	f := NewReservationFactory()
	orderID := uuid.New()

	t.Run("success", func(t *testing.T) {
		r, err := f.Create(ReserveInput{OrderID: orderID, Items: []ReservationItemInput{{ProductID: uuid.New(), Qty: 3}}})
		require.NoError(t, err)
		require.Len(t, r.Items, 1)
		assert.Equal(t, 3, r.Items[0].Qty.Int())
	})

	t.Run("empty items", func(t *testing.T) {
		_, err := f.Create(ReserveInput{OrderID: orderID})
		assert.ErrorIs(t, err, domain.ErrEmptyReservation)
	})

	t.Run("invalid quantity", func(t *testing.T) {
		_, err := f.Create(ReserveInput{OrderID: orderID, Items: []ReservationItemInput{{ProductID: uuid.New(), Qty: 0}}})
		assert.ErrorIs(t, err, domain.ErrInvalidQuantity)
	})

	t.Run("nil order id", func(t *testing.T) {
		_, err := f.Create(ReserveInput{OrderID: uuid.Nil, Items: []ReservationItemInput{{ProductID: uuid.New(), Qty: 1}}})
		assert.ErrorIs(t, err, domain.ErrInvalidOrderID)
	})

	t.Run("nil product id", func(t *testing.T) {
		_, err := f.Create(ReserveInput{OrderID: orderID, Items: []ReservationItemInput{{ProductID: uuid.Nil, Qty: 1}}})
		assert.ErrorIs(t, err, domain.ErrInvalidProductID)
	})
}
