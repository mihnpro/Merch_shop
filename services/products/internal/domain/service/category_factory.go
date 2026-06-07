package service

import (
	"context"

	"github.com/google/uuid"

	"github.com/mihnpro/Merch_shop/services/products/internal/domain"
	"github.com/mihnpro/Merch_shop/services/products/internal/domain/model"
	"github.com/mihnpro/Merch_shop/services/products/internal/domain/repository"
	vo "github.com/mihnpro/Merch_shop/services/products/internal/domain/valueobject"
)


type CreateCategoryInput struct {
	Code string
	Name string
}


type UpdateCategoryInput struct {
	ID     uuid.UUID
	Name   string
	Active bool
}


type CategoryFactory struct {
	categories repository.CategoryRepository
}

func NewCategoryFactory(categories repository.CategoryRepository) *CategoryFactory {
	return &CategoryFactory{categories: categories}
}


func (f *CategoryFactory) Create(ctx context.Context, in CreateCategoryInput) (*model.Category, error) {
	code, err := vo.NewCategoryCode(in.Code)
	if err != nil {
		return nil, err
	}
	exists, err := f.categories.ExistsByCode(ctx, code)
	if err != nil {
		return nil, err
	}
	if exists {
		return nil, domain.ErrCategoryAlreadyExists
	}
	return model.NewCategory(code, in.Name)
}

func (f *CategoryFactory) Update(ctx context.Context, in UpdateCategoryInput) (*model.Category, error) {
	category, err := f.categories.GetByID(ctx, in.ID)
	if err != nil {
		return nil, err
	}
	if err := category.Rename(in.Name); err != nil {
		return nil, err
	}
	category.SetActive(in.Active)
	return category, nil
}
