package repository

import (
	"context"

	"github.com/google/uuid"

	"github.com/mihnpro/Merch_shop/services/products/internal/domain/model"
	vo "github.com/mihnpro/Merch_shop/services/products/internal/domain/valueobject"
)

type CategoryRepository interface {
	Create(ctx context.Context, c *model.Category) (*model.Category, error)

	Update(ctx context.Context, c *model.Category) (*model.Category, error)

	GetByID(ctx context.Context, id uuid.UUID) (*model.Category, error)

	List(ctx context.Context, activeOnly bool) ([]*model.Category, error)

	ExistsByCode(ctx context.Context, code vo.CategoryCode) (bool, error)
}
