package service

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
	byID        map[uuid.UUID]*model.Product
	createErr   error
	updateErr   error
	deactErr    error
	listErr     error
	page        repository.ProductPage
	lastFilter  repository.ProductFilter
	created     *model.Product
	lastVersion int
	deactivated []uuid.UUID
}

func newFakeProductRepo() *fakeProductRepo {
	return &fakeProductRepo{byID: make(map[uuid.UUID]*model.Product)}
}

func (f *fakeProductRepo) Create(_ context.Context, p *model.Product) (*model.Product, error) {
	if f.createErr != nil {
		return nil, f.createErr
	}
	if p.ID == uuid.Nil {
		p.ID = uuid.New()
	}
	f.created = p
	f.byID[p.ID] = p
	return p, nil
}

func (f *fakeProductRepo) GetByID(_ context.Context, id uuid.UUID) (*model.Product, error) {
	if p, ok := f.byID[id]; ok {
		return p, nil
	}
	return nil, domain.ErrProductNotFound
}

func (f *fakeProductRepo) Update(_ context.Context, p *model.Product, expectedVersion int) (*model.Product, error) {
	if f.updateErr != nil {
		return nil, f.updateErr
	}
	f.lastVersion = expectedVersion
	f.byID[p.ID] = p
	return p, nil
}

func (f *fakeProductRepo) Deactivate(_ context.Context, id uuid.UUID) error {
	if f.deactErr != nil {
		return f.deactErr
	}
	f.deactivated = append(f.deactivated, id)
	return nil
}

func (f *fakeProductRepo) List(_ context.Context, filter repository.ProductFilter) (repository.ProductPage, error) {
	f.lastFilter = filter
	if f.listErr != nil {
		return repository.ProductPage{}, f.listErr
	}
	return f.page, nil
}

type fakeCategoryRepo struct {
	byID      map[uuid.UUID]*model.Category
	existsRet bool
	existsErr error
	createErr error
	updateErr error
	listErr   error
	list      []*model.Category
	created   *model.Category
}

func newFakeCategoryRepo() *fakeCategoryRepo {
	return &fakeCategoryRepo{byID: make(map[uuid.UUID]*model.Category)}
}

func (f *fakeCategoryRepo) put(c *model.Category) *model.Category {
	if c.ID == uuid.Nil {
		c.ID = uuid.New()
	}
	f.byID[c.ID] = c
	return c
}

func (f *fakeCategoryRepo) Create(_ context.Context, c *model.Category) (*model.Category, error) {
	if f.createErr != nil {
		return nil, f.createErr
	}
	f.created = c
	return f.put(c), nil
}

func (f *fakeCategoryRepo) Update(_ context.Context, c *model.Category) (*model.Category, error) {
	if f.updateErr != nil {
		return nil, f.updateErr
	}
	f.byID[c.ID] = c
	return c, nil
}

func (f *fakeCategoryRepo) GetByID(_ context.Context, id uuid.UUID) (*model.Category, error) {
	if c, ok := f.byID[id]; ok {
		return c, nil
	}
	return nil, domain.ErrCategoryNotFound
}

func (f *fakeCategoryRepo) List(_ context.Context, _ bool) ([]*model.Category, error) {
	if f.listErr != nil {
		return nil, f.listErr
	}
	return f.list, nil
}

func (f *fakeCategoryRepo) ExistsByCode(_ context.Context, _ vo.CategoryCode) (bool, error) {
	if f.existsErr != nil {
		return false, f.existsErr
	}
	return f.existsRet, nil
}

func activeCategory(id uuid.UUID) *model.Category {
	return &model.Category{
		ID:     id,
		Code:   vo.NewCategoryCodeFromStored("clothing"),
		Name:   "Clothes",
		Active: true,
	}
}
