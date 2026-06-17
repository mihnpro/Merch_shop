package query

import (
	"context"
	"errors"

	"github.com/google/uuid"

	"github.com/mihnpro/Merch_shop/services/products/internal/domain"
	"github.com/mihnpro/Merch_shop/services/products/internal/domain/model"
	"github.com/mihnpro/Merch_shop/services/products/internal/domain/repository"
	vo "github.com/mihnpro/Merch_shop/services/products/internal/domain/valueobject"
)

var errBoom = errors.New("boom")

type fakeProductRepo struct {
	byID       map[uuid.UUID]*model.Product
	getErr     error
	listErr    error
	page       repository.ProductPage
	lastFilter repository.ProductFilter
}

func newFakeProductRepo() *fakeProductRepo {
	return &fakeProductRepo{byID: make(map[uuid.UUID]*model.Product)}
}

func (f *fakeProductRepo) Create(_ context.Context, p *model.Product) (*model.Product, error) {
	return p, nil
}

func (f *fakeProductRepo) GetByID(_ context.Context, id uuid.UUID) (*model.Product, error) {
	if f.getErr != nil {
		return nil, f.getErr
	}
	if p, ok := f.byID[id]; ok {
		return p, nil
	}
	return nil, domain.ErrProductNotFound
}

func (f *fakeProductRepo) Update(_ context.Context, p *model.Product, _ int) (*model.Product, error) {
	return p, nil
}

func (f *fakeProductRepo) Deactivate(_ context.Context, _ uuid.UUID) error { return nil }

func (f *fakeProductRepo) List(_ context.Context, filter repository.ProductFilter) (repository.ProductPage, error) {
	f.lastFilter = filter
	if f.listErr != nil {
		return repository.ProductPage{}, f.listErr
	}
	return f.page, nil
}

type fakeCategoryRepo struct {
	listErr        error
	list           []*model.Category
	lastActiveOnly bool
}

func newFakeCategoryRepo() *fakeCategoryRepo { return &fakeCategoryRepo{} }

func (f *fakeCategoryRepo) Create(_ context.Context, c *model.Category) (*model.Category, error) {
	return c, nil
}

func (f *fakeCategoryRepo) Update(_ context.Context, c *model.Category) (*model.Category, error) {
	return c, nil
}

func (f *fakeCategoryRepo) GetByID(_ context.Context, _ uuid.UUID) (*model.Category, error) {
	return nil, domain.ErrCategoryNotFound
}

func (f *fakeCategoryRepo) List(_ context.Context, activeOnly bool) ([]*model.Category, error) {
	f.lastActiveOnly = activeOnly
	if f.listErr != nil {
		return nil, f.listErr
	}
	return f.list, nil
}

func (f *fakeCategoryRepo) ExistsByCode(_ context.Context, _ vo.CategoryCode) (bool, error) {
	return false, nil
}

func activeCategory(id uuid.UUID) *model.Category {
	return &model.Category{
		ID:     id,
		Code:   vo.NewCategoryCodeFromStored("clothing"),
		Name:   "Clothes",
		Active: true,
	}
}

func sampleProduct(id, catID uuid.UUID) *model.Product {
	p, _ := model.NewProduct("Shirt", "desc", vo.NewPricePointsFromStored(100), *activeCategory(catID), nil)
	p.ID = id
	return p
}
