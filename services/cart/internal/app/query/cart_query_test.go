package query

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"

	"github.com/mihnpro/Merch_shop/services/cart/internal/app/dto"
	"github.com/mihnpro/Merch_shop/services/cart/internal/domain"
	"github.com/mihnpro/Merch_shop/services/cart/internal/domain/model"
)

func TestGetCart(t *testing.T) {
	ctx := context.Background()

	t.Run("invalid user id", func(t *testing.T) {
		svc := NewCartQueryService(newFakeCartRepo(), &fakeCache{}, zap.NewNop())
		_, err := svc.GetCart(ctx, dto.GetCartInput{UserID: "bad"})
		assert.ErrorIs(t, err, domain.ErrInvalidUserID)
	})

	t.Run("cache hit skips database", func(t *testing.T) {
		repo := newFakeCartRepo()
		cache := &fakeCache{hit: true, items: []*model.CartItem{item(2, 100)}}
		svc := NewCartQueryService(repo, cache, zap.NewNop())
		view, err := svc.GetCart(ctx, dto.GetCartInput{UserID: uuid.NewString()})
		require.NoError(t, err)
		assert.Equal(t, 200, view.Total)
		assert.False(t, repo.getCartCalled, "DB must not be hit on cache hit")
	})

	t.Run("cache miss reads db and populates cache", func(t *testing.T) {
		repo := newFakeCartRepo()
		repo.items = []*model.CartItem{item(2, 100), item(1, 50)}
		cache := &fakeCache{hit: false}
		svc := NewCartQueryService(repo, cache, zap.NewNop())
		view, err := svc.GetCart(ctx, dto.GetCartInput{UserID: uuid.NewString()})
		require.NoError(t, err)
		assert.Equal(t, 250, view.Total)
		assert.Equal(t, 2, view.ItemCount)
		assert.True(t, repo.getCartCalled)
		assert.True(t, cache.setCalled)
	})

	t.Run("cart not found returns empty view", func(t *testing.T) {
		repo := newFakeCartRepo()
		repo.getCartErr = domain.ErrCartNotFound
		svc := NewCartQueryService(repo, &fakeCache{}, zap.NewNop())
		view, err := svc.GetCart(ctx, dto.GetCartInput{UserID: uuid.NewString()})
		require.NoError(t, err)
		assert.Equal(t, 0, view.Total)
		assert.Equal(t, 0, view.ItemCount)
		assert.NotNil(t, view.Items)
	})
}
