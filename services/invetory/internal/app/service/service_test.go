package service

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"

	"github.com/mihnpro/Merch_shop/services/invetory/internal/app/dto"
	"github.com/mihnpro/Merch_shop/services/invetory/internal/domain"
)

type deps struct {
	stock   *fakeStockRepo
	reserv  *fakeReservationRepo
	orders  *fakeOrderNotifier
	service InventoryService
}

func newDeps() *deps {
	stock := newFakeStockRepo()
	reserv := &fakeReservationRepo{}
	orders := &fakeOrderNotifier{}
	return &deps{
		stock:   stock,
		reserv:  reserv,
		orders:  orders,
		service: NewInventoryService(stock, reserv, orders, zap.NewNop()),
	}
}

func validReserveInput(orderID string) dto.ReserveStockInput {
	return dto.ReserveStockInput{
		OrderID: orderID,
		Items:   []dto.ReserveItemInput{{ProductID: uuid.NewString(), Qty: 2}},
	}
}

func TestReserveStock(t *testing.T) {
	ctx := context.Background()

	t.Run("success", func(t *testing.T) {
		d := newDeps()
		orderID := uuid.New()
		d.reserv.reserveResult = sampleReservation(orderID)
		view, err := d.service.ReserveStock(ctx, validReserveInput(orderID.String()))
		require.NoError(t, err)
		assert.Equal(t, orderID.String(), view.OrderID)
	})

	t.Run("invalid order id", func(t *testing.T) {
		d := newDeps()
		_, err := d.service.ReserveStock(ctx, validReserveInput("bad"))
		assert.ErrorIs(t, err, domain.ErrInvalidOrderID)
	})

	t.Run("invalid product id", func(t *testing.T) {
		d := newDeps()
		in := dto.ReserveStockInput{OrderID: uuid.NewString(), Items: []dto.ReserveItemInput{{ProductID: "bad", Qty: 1}}}
		_, err := d.service.ReserveStock(ctx, in)
		assert.ErrorIs(t, err, domain.ErrInvalidProductID)
	})

	t.Run("factory rejects bad quantity", func(t *testing.T) {
		d := newDeps()
		in := dto.ReserveStockInput{OrderID: uuid.NewString(), Items: []dto.ReserveItemInput{{ProductID: uuid.NewString(), Qty: 0}}}
		_, err := d.service.ReserveStock(ctx, in)
		assert.ErrorIs(t, err, domain.ErrInvalidQuantity)
	})

	t.Run("repository insufficient stock", func(t *testing.T) {
		d := newDeps()
		d.reserv.reserveErr = domain.ErrInsufficientStock
		_, err := d.service.ReserveStock(ctx, validReserveInput(uuid.NewString()))
		assert.ErrorIs(t, err, domain.ErrInsufficientStock)
	})
}

func TestReleaseReserve(t *testing.T) {
	ctx := context.Background()

	t.Run("success", func(t *testing.T) {
		d := newDeps()
		orderID := uuid.New()
		d.reserv.releaseResult = sampleReservation(orderID)
		view, err := d.service.ReleaseReserve(ctx, dto.ReleaseReserveInput{OrderID: orderID.String(), Reason: "cancelled"})
		require.NoError(t, err)
		assert.Equal(t, orderID.String(), view.OrderID)
	})

	t.Run("invalid order id", func(t *testing.T) {
		d := newDeps()
		_, err := d.service.ReleaseReserve(ctx, dto.ReleaseReserveInput{OrderID: "bad"})
		assert.ErrorIs(t, err, domain.ErrInvalidOrderID)
	})

	t.Run("not found", func(t *testing.T) {
		d := newDeps()
		d.reserv.releaseErr = domain.ErrReservationNotFound
		_, err := d.service.ReleaseReserve(ctx, dto.ReleaseReserveInput{OrderID: uuid.NewString()})
		assert.ErrorIs(t, err, domain.ErrReservationNotFound)
	})
}

func TestAdjustStock(t *testing.T) {
	ctx := context.Background()

	validAdjust := func() dto.AdjustStockInput {
		return dto.AdjustStockInput{ProductID: uuid.NewString(), Delta: 10, OperationID: uuid.NewString(), Reason: "restock"}
	}

	t.Run("success", func(t *testing.T) {
		d := newDeps()
		view, err := d.service.AdjustStock(ctx, validAdjust())
		require.NoError(t, err)
		assert.Equal(t, 10, view.Available)
	})

	t.Run("invalid product id", func(t *testing.T) {
		d := newDeps()
		in := validAdjust()
		in.ProductID = "bad"
		_, err := d.service.AdjustStock(ctx, in)
		assert.ErrorIs(t, err, domain.ErrInvalidProductID)
	})

	t.Run("invalid operation id", func(t *testing.T) {
		d := newDeps()
		in := validAdjust()
		in.OperationID = "bad"
		_, err := d.service.AdjustStock(ctx, in)
		assert.ErrorIs(t, err, domain.ErrInvalidOperationID)
	})

	t.Run("factory rejects zero delta", func(t *testing.T) {
		d := newDeps()
		in := validAdjust()
		in.Delta = 0
		_, err := d.service.AdjustStock(ctx, in)
		assert.ErrorIs(t, err, domain.ErrZeroDelta)
	})

	t.Run("repository insufficient stock", func(t *testing.T) {
		d := newDeps()
		d.stock.adjustErr = domain.ErrInsufficientStock
		_, err := d.service.AdjustStock(ctx, validAdjust())
		assert.ErrorIs(t, err, domain.ErrInsufficientStock)
	})
}

func TestHandleOrderCreated(t *testing.T) {
	ctx := context.Background()

	t.Run("reserve ok confirms order", func(t *testing.T) {
		d := newDeps()
		orderID := uuid.New()
		d.reserv.reserveResult = sampleReservation(orderID)
		require.NoError(t, d.service.HandleOrderCreated(ctx, validReserveInput(orderID.String())))
		assert.Equal(t, 1, d.orders.calls)
		assert.Equal(t, "confirmed", d.orders.lastStatus)
	})

	t.Run("reserve failure cancels order", func(t *testing.T) {
		d := newDeps()
		d.reserv.reserveErr = domain.ErrInsufficientStock
		require.NoError(t, d.service.HandleOrderCreated(ctx, validReserveInput(uuid.NewString())))
		assert.Equal(t, "cancelled", d.orders.lastStatus)
	})

	t.Run("propagates notifier error", func(t *testing.T) {
		d := newDeps()
		orderID := uuid.New()
		d.reserv.reserveResult = sampleReservation(orderID)
		d.orders.err = errBoom
		err := d.service.HandleOrderCreated(ctx, validReserveInput(orderID.String()))
		assert.ErrorIs(t, err, errBoom)
	})
}
