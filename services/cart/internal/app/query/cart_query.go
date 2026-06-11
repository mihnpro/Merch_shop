package query

import (
	"context"

	"github.com/google/uuid"
	"go.uber.org/zap"

	"github.com/mihnpro/Merch_shop/services/cart/internal/app/dto"
	"github.com/mihnpro/Merch_shop/services/cart/internal/domain"
	"github.com/mihnpro/Merch_shop/services/cart/internal/domain/repository"
)

type CartQueryService interface {
	GetCart(ctx context.Context, in dto.GetCartInput) (dto.CartView, error)
}

type cartQueryService struct {
	repo  repository.CartRepository
	cache repository.CacheRepository
	log   *zap.Logger
}

func NewCartQueryService(
	repo repository.CartRepository,
	cache repository.CacheRepository,
	log *zap.Logger,
) CartQueryService {
	return &cartQueryService{repo: repo, cache: cache, log: log}
}

func (s *cartQueryService) GetCart(ctx context.Context, in dto.GetCartInput) (dto.CartView, error) {
	userID, err := uuid.Parse(in.UserID)
	if err != nil {
		return dto.CartView{}, domain.ErrInvalidUserID
	}

	if s.cache != nil {
		if items, ok := s.cache.TryFromCache(ctx, in.UserID); ok {
			return dto.ToCartView(items), nil
		}
	}

	cart, err := s.repo.GetCartByUserID(ctx, userID)
	if err != nil {
		if err == domain.ErrCartNotFound {
			return dto.CartView{Items: []dto.CartItemView{}, Total: 0, ItemCount: 0}, nil
		}
		return dto.CartView{}, err
	}

	items, err := s.repo.GetCartItems(ctx, cart.ID)
	if err != nil {
		return dto.CartView{}, err
	}

	if s.cache != nil {
		s.cache.SetCache(ctx, in.UserID, items)
	}

	return dto.ToCartView(items), nil
}
