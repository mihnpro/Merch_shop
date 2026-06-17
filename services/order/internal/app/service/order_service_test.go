package service

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/mihnpro/Merch_shop/services/order/internal/app/dto"
	"github.com/mihnpro/Merch_shop/services/order/internal/app/port"
	"github.com/mihnpro/Merch_shop/services/order/internal/domain"
	"github.com/mihnpro/Merch_shop/services/order/internal/domain/events"
	"github.com/mihnpro/Merch_shop/services/order/internal/domain/model"
)

type deps struct {
	orders    *fakeOrderRepo
	cart      *fakeCart
	user      *fakeUser
	inventory *fakeInventory
	svc       OrderService
}

func newDeps() *deps {
	orders := newFakeOrderRepo()
	cart := &fakeCart{}
	user := &fakeUser{}
	inventory := &fakeInventory{}
	return &deps{
		orders:    orders,
		cart:      cart,
		user:      user,
		inventory: inventory,
		svc:       NewOrderService(orders, cart, user, inventory),
	}
}

func createInput(userID string) dto.CreateOrderInput {
	return dto.CreateOrderInput{UserID: userID, DeliveryAddress: "123 Main St"}
}

func TestCreateOrder(t *testing.T) {
	ctx := context.Background()
	userID := uuid.New()

	t.Run("success deducts points and clears cart", func(t *testing.T) {
		d := newDeps()
		productID := uuid.New()
		d.cart.items = []port.CartItem{cartItem(productID, 2, 100)}
		d.user.balance = 500
		view, err := d.svc.CreateOrder(ctx, createInput(userID.String()))
		require.NoError(t, err)
		assert.Equal(t, string(model.StatusPending), view.Status)
		assert.Equal(t, int64(200), view.TotalPoints)
		assert.Len(t, view.Items, 1)
		assert.Equal(t, 1, d.user.deductCalls)
		assert.Equal(t, int64(200), d.user.deductAmount)
		assert.True(t, d.cart.clearCalled)
		assert.NotNil(t, d.orders.createdOrder)

		var ev events.OrderCreated
		require.NoError(t, json.Unmarshal(d.orders.createdPayld, &ev))
		assert.Equal(t, view.ID, ev.OrderID)
		assert.Equal(t, userID.String(), ev.UserID)
		require.Len(t, ev.Items, 1)
		assert.Equal(t, productID.String(), ev.Items[0].ProductID)
		assert.Equal(t, 2, ev.Items[0].Quantity)
	})

	t.Run("invalid user id", func(t *testing.T) {
		d := newDeps()
		_, err := d.svc.CreateOrder(ctx, createInput("bad"))
		assert.ErrorIs(t, err, domain.ErrInvalidInput)
	})

	t.Run("cart fetch error propagates", func(t *testing.T) {
		d := newDeps()
		d.cart.getErr = errBoom
		_, err := d.svc.CreateOrder(ctx, createInput(userID.String()))
		assert.ErrorIs(t, err, errBoom)
	})

	t.Run("empty cart", func(t *testing.T) {
		d := newDeps()
		d.cart.items = nil
		_, err := d.svc.CreateOrder(ctx, createInput(userID.String()))
		assert.ErrorIs(t, err, domain.ErrEmptyCart)
	})

	t.Run("malformed product id in cart", func(t *testing.T) {
		d := newDeps()
		d.cart.items = []port.CartItem{{ProductID: "bad", Quantity: 1, PricePoints: 10}}
		_, err := d.svc.CreateOrder(ctx, createInput(userID.String()))
		assert.ErrorIs(t, err, domain.ErrInvalidInput)
	})

	t.Run("insufficient balance does not deduct", func(t *testing.T) {
		d := newDeps()
		d.cart.items = []port.CartItem{cartItem(uuid.New(), 2, 100)}
		d.user.balance = 199
		_, err := d.svc.CreateOrder(ctx, createInput(userID.String()))
		assert.ErrorIs(t, err, domain.ErrInsufficientBalance)
		assert.Equal(t, 0, d.user.deductCalls)
	})

	t.Run("deduct failure aborts before save", func(t *testing.T) {
		d := newDeps()
		d.cart.items = []port.CartItem{cartItem(uuid.New(), 1, 100)}
		d.user.balance = 500
		d.user.deductErr = errBoom
		_, err := d.svc.CreateOrder(ctx, createInput(userID.String()))
		assert.ErrorIs(t, err, errBoom)
		assert.Nil(t, d.orders.createdOrder)
	})

	t.Run("save failure refunds points", func(t *testing.T) {
		d := newDeps()
		d.cart.items = []port.CartItem{cartItem(uuid.New(), 1, 100)}
		d.user.balance = 500
		d.orders.createErr = errBoom
		_, err := d.svc.CreateOrder(ctx, createInput(userID.String()))
		assert.ErrorIs(t, err, errBoom)
		assert.Equal(t, 1, d.user.addCalls)
		assert.Equal(t, int64(100), d.user.addAmount)
	})
}

func TestUpdateOrderStatus(t *testing.T) {
	ctx := context.Background()

	seedOrder := func(d *deps, status model.OrderStatus) *model.Order {
		o, err := model.NewOrder(uuid.New(), []model.OrderItem{
			{ProductID: uuid.New(), ProductName: "Item", Quantity: 1, PricePoints: 100},
		}, "addr", nil)
		require.NoError(t, err)
		o.Status = status
		d.orders.byID[o.ID] = o
		return o
	}

	t.Run("invalid order id", func(t *testing.T) {
		d := newDeps()
		_, err := d.svc.UpdateOrderStatus(ctx, dto.UpdateOrderStatusInput{OrderID: "bad", NewStatus: "confirmed"})
		assert.ErrorIs(t, err, domain.ErrInvalidInput)
	})

	t.Run("order not found", func(t *testing.T) {
		d := newDeps()
		_, err := d.svc.UpdateOrderStatus(ctx, dto.UpdateOrderStatusInput{OrderID: uuid.NewString(), NewStatus: "confirmed"})
		assert.ErrorIs(t, err, domain.ErrOrderNotFound)
	})

	t.Run("same status is idempotent", func(t *testing.T) {
		d := newDeps()
		o := seedOrder(d, model.StatusPending)
		_, err := d.svc.UpdateOrderStatus(ctx, dto.UpdateOrderStatusInput{OrderID: o.ID.String(), NewStatus: "pending"})
		require.NoError(t, err)
		assert.Equal(t, 0, d.orders.updateCalls)
	})

	t.Run("invalid transition", func(t *testing.T) {
		d := newDeps()
		o := seedOrder(d, model.StatusPending)
		_, err := d.svc.UpdateOrderStatus(ctx, dto.UpdateOrderStatusInput{OrderID: o.ID.String(), NewStatus: "delivered"})
		assert.ErrorIs(t, err, domain.ErrInvalidStatusChange)
	})

	t.Run("valid transition persists", func(t *testing.T) {
		d := newDeps()
		o := seedOrder(d, model.StatusPending)
		view, err := d.svc.UpdateOrderStatus(ctx, dto.UpdateOrderStatusInput{OrderID: o.ID.String(), NewStatus: "confirmed", Reason: "ok"})
		require.NoError(t, err)
		assert.Equal(t, "confirmed", view.Status)
		assert.Equal(t, model.StatusConfirmed, d.orders.updatedTo)
	})

	t.Run("cancel refunds points and releases reserve", func(t *testing.T) {
		d := newDeps()
		o := seedOrder(d, model.StatusPending)
		_, err := d.svc.UpdateOrderStatus(ctx, dto.UpdateOrderStatusInput{OrderID: o.ID.String(), NewStatus: "cancelled", Reason: "changed mind"})
		require.NoError(t, err)
		assert.Equal(t, 1, d.user.addCalls)
		assert.Equal(t, o.TotalPoints, d.user.addAmount)
		assert.True(t, d.inventory.releaseCalled)
	})
}
