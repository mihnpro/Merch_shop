package service

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"go.uber.org/zap"

	"github.com/mihnpro/Merch_shop/services/cart/internal/app/dto"
	"github.com/mihnpro/Merch_shop/services/cart/internal/app/port"
	"github.com/mihnpro/Merch_shop/services/cart/internal/domain"
	"github.com/mihnpro/Merch_shop/services/cart/internal/domain/repository"
	domainsvc "github.com/mihnpro/Merch_shop/services/cart/internal/domain/service"
	vo "github.com/mihnpro/Merch_shop/services/cart/internal/domain/valueobject"
)

type CartService interface {
	AddItem(ctx context.Context, in dto.AddItemInput) (dto.CartItemView, error)
	UpdateItem(ctx context.Context, in dto.UpdateItemInput) (dto.CartItemView, error)
	RemoveItem(ctx context.Context, in dto.RemoveItemInput) error
	ClearCart(ctx context.Context, in dto.ClearCartInput) error
}

type cartService struct {
	psqlrepo  repository.CartRepository
	products  port.ProductAdapter
	inventory port.InventoryAdapter
	factory   *domainsvc.CartFactory
	cache     repository.CacheRepository
	log       *zap.Logger
}

func NewCartService(
	repo repository.CartRepository,
	products port.ProductAdapter,
	inventory port.InventoryAdapter,
	cache repository.CacheRepository,
	log *zap.Logger,
) CartService {
	return &cartService{
		psqlrepo:  repo,
		products:  products,
		inventory: inventory,
		factory:   domainsvc.NewCartFactory(),
		cache:     cache,
		log:       log,
	}
}

func (s *cartService) AddItem(ctx context.Context, in dto.AddItemInput) (dto.CartItemView, error) {
	if in.UserID == "" || in.ProductID == "" {
		return dto.CartItemView{}, domain.ErrInvalidInput
	}

	userID, err := uuid.Parse(in.UserID)
	if err != nil {
		return dto.CartItemView{}, domain.ErrInvalidUserID
	}
	productID, err := uuid.Parse(in.ProductID)
	if err != nil {
		return dto.CartItemView{}, domain.ErrInvalidProductID
	}

	inStock, err := s.inventory.CheckStock(ctx, in.ProductID, in.Quantity)
	if err != nil {
		return dto.CartItemView{}, err
	}
	if !inStock {
		return dto.CartItemView{}, domain.ErrOutOfStock
	}

	product, err := s.products.GetProduct(ctx, in.ProductID)
	if err != nil {
		return dto.CartItemView{}, err
	}
	if !product.Active {
		return dto.CartItemView{}, domain.ErrProductInactive
	}

	cart, err := s.psqlrepo.GetOrCreateCart(ctx, userID)
	if err != nil {
		return dto.CartItemView{}, err
	}

	existing, err := s.psqlrepo.GetItemByProduct(ctx, cart.ID, productID)
	if err != nil && !errors.Is(err, domain.ErrItemNotFound) {
		return dto.CartItemView{}, err
	}

	var resultItem dto.CartItemView

	if existing != nil {
		newQty := existing.Quantity.Int() + in.Quantity
		inStock, err = s.inventory.CheckStock(ctx, in.ProductID, newQty)
		if err != nil {
			return dto.CartItemView{}, err
		}
		if !inStock {
			return dto.CartItemView{}, domain.ErrOutOfStock
		}
		if err := s.psqlrepo.UpdateItemQty(ctx, existing.ID, newQty); err != nil {
			return dto.CartItemView{}, err
		}
		existing.Quantity = vo.NewQuantityFromStored(newQty)
		resultItem = dto.ToCartItemView(existing)
	} else {
		newItem, err := s.factory.NewCartItem(domainsvc.NewCartItemInput{
			CartID:      cart.ID,
			ProductID:   productID,
			Quantity:    in.Quantity,
			PriceAtAdd:  product.PricePoints,
			ProductName: product.Name,
		})
		if err != nil {
			return dto.CartItemView{}, err
		}
		saved, err := s.psqlrepo.InsertItem(ctx, newItem)
		if err != nil {
			return dto.CartItemView{}, err
		}
		resultItem = dto.ToCartItemView(saved)
	}

	_ = s.psqlrepo.TouchCart(ctx, cart.ID)
	s.cache.InvalidateCache(ctx, in.UserID)

	return resultItem, nil
}

func (s *cartService) UpdateItem(ctx context.Context, in dto.UpdateItemInput) (dto.CartItemView, error) {
	if in.UserID == "" || in.ItemID == "" {
		return dto.CartItemView{}, domain.ErrInvalidInput
	}

	if in.NewQuantity < 0 {
		return dto.CartItemView{}, domain.ErrInvalidQuantity
	}

	if in.NewQuantity == 0 {
		err := s.RemoveItem(ctx, dto.RemoveItemInput{UserID: in.UserID, ItemID: in.ItemID})
		return dto.CartItemView{}, err
	}

	userID, err := uuid.Parse(in.UserID)
	if err != nil {
		return dto.CartItemView{}, domain.ErrInvalidUserID
	}
	itemID, err := uuid.Parse(in.ItemID)
	if err != nil {
		return dto.CartItemView{}, domain.ErrInvalidItemID
	}

	item, err := s.psqlrepo.GetItemByID(ctx, itemID)
	if err != nil {
		return dto.CartItemView{}, err
	}

	cart, err := s.psqlrepo.GetCartByUserID(ctx, userID)
	if err != nil {
		return dto.CartItemView{}, err
	}
	if item.CartID != cart.ID {
		return dto.CartItemView{}, domain.ErrItemNotFound
	}

	if in.NewQuantity > item.Quantity.Int() {
		inStock, err := s.inventory.CheckStock(ctx, item.ProductID.String(), in.NewQuantity)
		if err != nil {
			return dto.CartItemView{}, err
		}
		if !inStock {
			return dto.CartItemView{}, domain.ErrOutOfStock
		}
	}

	if err := s.psqlrepo.UpdateItemQty(ctx, itemID, in.NewQuantity); err != nil {
		return dto.CartItemView{}, err
	}

	item.Quantity = vo.NewQuantityFromStored(in.NewQuantity)
	s.cache.InvalidateCache(ctx, in.UserID)

	return dto.ToCartItemView(item), nil
}

func (s *cartService) RemoveItem(ctx context.Context, in dto.RemoveItemInput) error {
	if in.UserID == "" || in.ItemID == "" {
		return domain.ErrInvalidInput
	}

	userID, err := uuid.Parse(in.UserID)
	if err != nil {
		return domain.ErrInvalidUserID
	}
	itemID, err := uuid.Parse(in.ItemID)
	if err != nil {
		return domain.ErrInvalidItemID
	}

	cart, err := s.psqlrepo.GetCartByUserID(ctx, userID)
	if err != nil {
		if errors.Is(err, domain.ErrCartNotFound) {
			return nil
		}
		return err
	}

	if err := s.psqlrepo.DeleteItem(ctx, itemID, cart.ID); err != nil {
		return err
	}

	s.cache.InvalidateCache(ctx, in.UserID)
	return nil
}

func (s *cartService) ClearCart(ctx context.Context, in dto.ClearCartInput) error {
	if in.UserID == "" {
		return domain.ErrInvalidInput
	}

	userID, err := uuid.Parse(in.UserID)
	if err != nil {
		return domain.ErrInvalidUserID
	}

	cart, err := s.psqlrepo.GetOrCreateCart(ctx, userID)
	if err != nil {
		return err
	}

	if err := s.psqlrepo.ClearCart(ctx, cart.ID); err != nil {
		return err
	}

	s.cache.InvalidateCache(ctx, in.UserID)
	return nil
}
