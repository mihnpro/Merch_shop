package query

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/mihnpro/Merch_shop/services/order/internal/app/dto"
	"github.com/mihnpro/Merch_shop/services/order/internal/domain"
	"github.com/mihnpro/Merch_shop/services/order/internal/domain/model"
	"github.com/mihnpro/Merch_shop/services/order/internal/domain/repository"
)

func TestGetOrder(t *testing.T) {
	ctx := context.Background()

	t.Run("success", func(t *testing.T) {
		repo := newFakeOrderRepo()
		o := sampleOrder(uuid.New())
		repo.byID[o.ID] = o
		view, err := NewOrderQueryService(repo).GetOrder(ctx, dto.GetOrderInput{OrderID: o.ID.String()})
		require.NoError(t, err)
		assert.Equal(t, o.ID.String(), view.ID)
	})

	t.Run("invalid id", func(t *testing.T) {
		_, err := NewOrderQueryService(newFakeOrderRepo()).GetOrder(ctx, dto.GetOrderInput{OrderID: "bad"})
		assert.ErrorIs(t, err, domain.ErrInvalidInput)
	})

	t.Run("not found", func(t *testing.T) {
		_, err := NewOrderQueryService(newFakeOrderRepo()).GetOrder(ctx, dto.GetOrderInput{OrderID: uuid.NewString()})
		assert.ErrorIs(t, err, domain.ErrOrderNotFound)
	})
}

func TestListOrders(t *testing.T) {
	ctx := context.Background()
	userID := uuid.New()

	t.Run("success maps and passes filter", func(t *testing.T) {
		repo := newFakeOrderRepo()
		repo.listResult = []*model.Order{sampleOrder(userID), sampleOrder(userID)}
		repo.listToken = "next"
		views, token, err := NewOrderQueryService(repo).ListOrders(ctx, dto.ListOrdersInput{
			UserID: userID.String(), Status: "pending", PageSize: 10,
		})
		require.NoError(t, err)
		assert.Len(t, views, 2)
		assert.Equal(t, "next", token)
		require.NotNil(t, repo.lastFilter.UserID)
		assert.Equal(t, userID, *repo.lastFilter.UserID)
		assert.Equal(t, "pending", repo.lastFilter.Status)
	})

	t.Run("empty user id leaves filter nil", func(t *testing.T) {
		repo := newFakeOrderRepo()
		_, _, err := NewOrderQueryService(repo).ListOrders(ctx, dto.ListOrdersInput{})
		require.NoError(t, err)
		assert.Nil(t, repo.lastFilter.UserID)
	})

	t.Run("invalid user id", func(t *testing.T) {
		_, _, err := NewOrderQueryService(newFakeOrderRepo()).ListOrders(ctx, dto.ListOrdersInput{UserID: "bad"})
		assert.ErrorIs(t, err, domain.ErrInvalidInput)
	})

	t.Run("repository error propagates", func(t *testing.T) {
		repo := newFakeOrderRepo()
		repo.listErr = errBoom
		_, _, err := NewOrderQueryService(repo).ListOrders(ctx, dto.ListOrdersInput{})
		assert.ErrorIs(t, err, errBoom)
	})
}

func TestGetAnalytics(t *testing.T) {
	ctx := context.Background()

	t.Run("computes average and maps top products", func(t *testing.T) {
		repo := newFakeOrderRepo()
		repo.analytics = repository.Analytics{
			OrdersCount: 4,
			PointsSpent: 1000,
			TopProducts: []repository.TopProduct{{ProductID: uuid.New(), ProductName: "Hoodie", Quantity: 7}},
		}
		view, err := NewOrderQueryService(repo).GetAnalytics(ctx, "week")
		require.NoError(t, err)
		assert.Equal(t, "week", view.Period)
		assert.Equal(t, int64(250), view.AverageOrderValue)
		require.Len(t, view.TopProducts, 1)
		assert.Equal(t, "Hoodie", view.TopProducts[0].ProductName)
	})

	t.Run("zero orders yields zero average", func(t *testing.T) {
		repo := newFakeOrderRepo()
		repo.analytics = repository.Analytics{OrdersCount: 0, PointsSpent: 0}
		view, err := NewOrderQueryService(repo).GetAnalytics(ctx, "unknown")
		require.NoError(t, err)
		assert.Equal(t, int64(0), view.AverageOrderValue)
		assert.Equal(t, "day", view.Period)
	})

	t.Run("month period normalized", func(t *testing.T) {
		repo := newFakeOrderRepo()
		view, err := NewOrderQueryService(repo).GetAnalytics(ctx, "month")
		require.NoError(t, err)
		assert.Equal(t, "month", view.Period)
	})

	t.Run("repository error propagates", func(t *testing.T) {
		repo := newFakeOrderRepo()
		repo.analyticsErr = errBoom
		_, err := NewOrderQueryService(repo).GetAnalytics(ctx, "day")
		assert.ErrorIs(t, err, errBoom)
	})
}
