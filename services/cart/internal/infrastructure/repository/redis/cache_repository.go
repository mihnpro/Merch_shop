package redis

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/mihnpro/Merch_shop/services/cart/internal/domain/model"
	"github.com/mihnpro/Merch_shop/services/cart/internal/domain/repository"
	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
)

const cacheTTL = 30 * time.Minute

type cacheRepository struct {
	client *redis.Client
	log    *zap.Logger
}

func NewCacheRepository(client *redis.Client, log *zap.Logger) repository.CacheRepository {
	return &cacheRepository{client: client, log: log}
}

func (s *cacheRepository) TryFromCache(ctx context.Context, userID string) ([]*model.CartItem, bool) {
	data, err := s.client.Get(ctx, cacheKey(userID)).Bytes()
	if err != nil {
		return nil, false
	}
	var items []*model.CartItem
	if err := json.Unmarshal(data, &items); err != nil {
		s.log.Warn("cache unmarshal failed", zap.Error(err))
		return nil, false
	}
	return items, true
}

func (s *cacheRepository) SetCache(ctx context.Context, userID string, items []*model.CartItem) {
	data, err := json.Marshal(items)
	if err != nil {
		s.log.Warn("cache marshal failed", zap.Error(err))
		return
	}
	if err := s.client.Set(ctx, cacheKey(userID), data, cacheTTL).Err(); err != nil {
		s.log.Warn("cache set failed", zap.Error(err))
	}
}

func (s *cacheRepository) InvalidateCache(ctx context.Context, userID string) {
	if s.client == nil {
		return
	}
	_ = s.client.Del(ctx, cacheKey(userID)).Err()
}

func cacheKey(userID string) string {
	return fmt.Sprintf("cart:%s", userID)
}
