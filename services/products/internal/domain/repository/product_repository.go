package repository

import (
	"context"
	"time"

	"github.com/google/uuid"

	"github.com/mihnpro/Merch_shop/services/products/internal/domain/model"
)

type ProductCursor struct {
	CreatedAt time.Time
	ID        uuid.UUID
}

type ProductFilter struct {
	CategoryID *uuid.UUID
	Search     string
	ActiveOnly bool
	Limit      int
	Cursor     *ProductCursor
}

type ProductPage struct {
	Products []*model.Product
	Next     *ProductCursor
}

type ProductRepository interface {
	Create(ctx context.Context, p *model.Product) (*model.Product, error)

	GetByID(ctx context.Context, id uuid.UUID) (*model.Product, error)

	Update(ctx context.Context, p *model.Product, expectedVersion int) (*model.Product, error)

	Deactivate(ctx context.Context, id uuid.UUID) error

	List(ctx context.Context, f ProductFilter) (ProductPage, error)
}
