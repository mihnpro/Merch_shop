package query

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/mihnpro/Merch_shop/services/invetory/internal/app/dto"
	"github.com/mihnpro/Merch_shop/services/invetory/internal/domain"
	"github.com/mihnpro/Merch_shop/services/invetory/internal/domain/model"
)

func TestCheckStock(t *testing.T) {
	ctx := context.Background()
	prodID := uuid.New()

	t.Run("invalid product id", func(t *testing.T) {
		svc := NewInventoryReadService(newFakeStockRepo())
		_, err := svc.CheckStock(ctx, dto.CheckStockInput{ProductID: "bad"})
		assert.ErrorIs(t, err, domain.ErrInvalidProductID)
	})

	t.Run("not found returns out of stock", func(t *testing.T) {
		svc := NewInventoryReadService(newFakeStockRepo())
		view, err := svc.CheckStock(ctx, dto.CheckStockInput{ProductID: uuid.NewString(), Qty: 1})
		require.NoError(t, err)
		assert.Equal(t, 0, view.Available)
		assert.False(t, view.InStock)
	})

	t.Run("in stock when available covers need", func(t *testing.T) {
		repo := newFakeStockRepo()
		repo.byID[prodID] = stockAt(prodID, 10)
		svc := NewInventoryReadService(repo)
		view, err := svc.CheckStock(ctx, dto.CheckStockInput{ProductID: prodID.String(), Qty: 5})
		require.NoError(t, err)
		assert.Equal(t, 10, view.Available)
		assert.True(t, view.InStock)
	})

	t.Run("out of stock when need exceeds available", func(t *testing.T) {
		repo := newFakeStockRepo()
		repo.byID[prodID] = stockAt(prodID, 3)
		svc := NewInventoryReadService(repo)
		view, err := svc.CheckStock(ctx, dto.CheckStockInput{ProductID: prodID.String(), Qty: 20})
		require.NoError(t, err)
		assert.False(t, view.InStock)
	})

	t.Run("need defaults to one when qty below one", func(t *testing.T) {
		repo := newFakeStockRepo()
		repo.byID[prodID] = stockAt(prodID, 1)
		svc := NewInventoryReadService(repo)
		view, err := svc.CheckStock(ctx, dto.CheckStockInput{ProductID: prodID.String(), Qty: 0})
		require.NoError(t, err)
		assert.True(t, view.InStock)
	})

	t.Run("propagates repository error", func(t *testing.T) {
		repo := newFakeStockRepo()
		repo.getErr = errBoom
		svc := NewInventoryReadService(repo)
		_, err := svc.CheckStock(ctx, dto.CheckStockInput{ProductID: prodID.String()})
		assert.ErrorIs(t, err, errBoom)
	})
}

func TestListStock(t *testing.T) {
	ctx := context.Background()

	t.Run("maps results", func(t *testing.T) {
		repo := newFakeStockRepo()
		repo.list = []*model.Stock{stockAt(uuid.New(), 5), stockAt(uuid.New(), 9)}
		svc := NewInventoryReadService(repo)
		views, err := svc.ListStock(ctx)
		require.NoError(t, err)
		assert.Len(t, views, 2)
	})

	t.Run("propagates repository error", func(t *testing.T) {
		repo := newFakeStockRepo()
		repo.listErr = errBoom
		svc := NewInventoryReadService(repo)
		_, err := svc.ListStock(ctx)
		assert.ErrorIs(t, err, errBoom)
	})
}

func TestListStockByIDs(t *testing.T) {
	ctx := context.Background()

	t.Run("skips invalid ids and queries valid ones", func(t *testing.T) {
		valid := uuid.New()
		repo := newFakeStockRepo()
		repo.many = []*model.Stock{stockAt(valid, 7)}
		svc := NewInventoryReadService(repo)
		views, err := svc.ListStockByIDs(ctx, []string{"bad", valid.String()})
		require.NoError(t, err)
		assert.Len(t, views, 1)
		assert.Equal(t, []uuid.UUID{valid}, repo.lastMany)
	})

	t.Run("all invalid returns empty without querying", func(t *testing.T) {
		repo := newFakeStockRepo()
		svc := NewInventoryReadService(repo)
		views, err := svc.ListStockByIDs(ctx, []string{"bad", "worse"})
		require.NoError(t, err)
		assert.Empty(t, views)
		assert.Nil(t, repo.lastMany)
	})

	t.Run("propagates repository error", func(t *testing.T) {
		repo := newFakeStockRepo()
		repo.getManyErr = errBoom
		svc := NewInventoryReadService(repo)
		_, err := svc.ListStockByIDs(ctx, []string{uuid.NewString()})
		assert.ErrorIs(t, err, errBoom)
	})
}
