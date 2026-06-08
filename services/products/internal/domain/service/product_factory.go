package service

import (
	"context"

	"github.com/google/uuid"

	"github.com/mihnpro/Merch_shop/services/products/internal/domain/model"
	"github.com/mihnpro/Merch_shop/services/products/internal/domain/repository"
	vo "github.com/mihnpro/Merch_shop/services/products/internal/domain/valueobject"
)

type CreateProductInput struct {
	Name        string
	Description string
	PricePoints int64
	CategoryID  uuid.UUID
	PhotoKeys   []string
}


type UpdateProductInput struct {
	ProductID   uuid.UUID
	Name        string
	Description string
	PricePoints int64
	CategoryID  uuid.UUID
	PhotoKeys   []string
	Active      bool
	Version     int
}

type ProductFactory struct {
	products   repository.ProductRepository
	categories repository.CategoryRepository
}

func NewProductFactory(products repository.ProductRepository, categories repository.CategoryRepository) *ProductFactory {
	return &ProductFactory{products: products, categories: categories}
}

func (f *ProductFactory) Create(ctx context.Context, in CreateProductInput) (*model.Product, error) {
	price, err := vo.NewPricePoints(in.PricePoints)
	if err != nil {
		return nil, err
	}
	photos, err := buildPhotoKeys(in.PhotoKeys)
	if err != nil {
		return nil, err
	}
	category, err := f.categories.GetByID(ctx, in.CategoryID)
	if err != nil {
		return nil, err
	}
	return model.NewProduct(in.Name, in.Description, price, *category, photos)
}


func (f *ProductFactory) Update(ctx context.Context, in UpdateProductInput) (*model.Product, error) {
	product, err := f.products.GetByID(ctx, in.ProductID)
	if err != nil {
		return nil, err
	}
	price, err := vo.NewPricePoints(in.PricePoints)
	if err != nil {
		return nil, err
	}
	photos, err := buildPhotoKeys(in.PhotoKeys)
	if err != nil {
		return nil, err
	}
	category, err := f.categories.GetByID(ctx, in.CategoryID)
	if err != nil {
		return nil, err
	}
	if err := product.ApplyUpdate(in.Name, in.Description, price, *category, photos, in.Active, in.Version); err != nil {
		return nil, err
	}
	return product, nil
}


func buildPhotoKeys(raw []string) ([]vo.PhotoKey, error) {
	if len(raw) == 0 {
		return nil, nil
	}
	keys := make([]vo.PhotoKey, 0, len(raw))
	for _, r := range raw {
		key, err := vo.NewPhotoKey(r)
		if err != nil {
			return nil, err
		}
		if key.IsEmpty() {
			continue
		}
		keys = append(keys, key)
	}
	return keys, nil
}
