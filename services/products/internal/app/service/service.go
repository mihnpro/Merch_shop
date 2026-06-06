package service

import (
	"context"

	"github.com/google/uuid"

	"github.com/mihnpro/Merch_shop/services/products/internal/app/dto"
	"github.com/mihnpro/Merch_shop/services/products/internal/domain"
	"github.com/mihnpro/Merch_shop/services/products/internal/domain/repository"
	domainsvc "github.com/mihnpro/Merch_shop/services/products/internal/domain/service"
)

type CatalogService interface {
	CreateProduct(ctx context.Context, in dto.CreateProductInput) (dto.ProductView, error)
	UpdateProduct(ctx context.Context, in dto.UpdateProductInput) (dto.ProductView, error)
	DeactivateProduct(ctx context.Context, productID string) error
	CreateCategory(ctx context.Context, in dto.CreateCategoryInput) (dto.CategoryView, error)
	UpdateCategory(ctx context.Context, in dto.UpdateCategoryInput) (dto.CategoryView, error)
}

type catalogService struct {
	products        repository.ProductRepository
	categories      repository.CategoryRepository
	productFactory  *domainsvc.ProductFactory
	categoryFactory *domainsvc.CategoryFactory
}

func NewCatalogService(products repository.ProductRepository, categories repository.CategoryRepository) CatalogService {
	return &catalogService{
		products:        products,
		categories:      categories,
		productFactory:  domainsvc.NewProductFactory(products, categories),
		categoryFactory: domainsvc.NewCategoryFactory(categories),
	}
}

func (s *catalogService) CreateProduct(ctx context.Context, in dto.CreateProductInput) (dto.ProductView, error) {
	categoryID, err := parseID(in.CategoryID, domain.ErrCategoryNotFound)
	if err != nil {
		return dto.ProductView{}, err
	}

	product, err := s.productFactory.Create(ctx, domainsvc.CreateProductInput{
		Name:        in.Name,
		Description: in.Description,
		PricePoints: in.PricePoints,
		CategoryID:  categoryID,
		Sizes:       in.Sizes,
		PhotoKeys:   in.PhotoKeys,
	})
	if err != nil {
		return dto.ProductView{}, err
	}

	saved, err := s.products.Create(ctx, product)
	if err != nil {
		return dto.ProductView{}, err
	}
	return dto.ToProductView(saved), nil
}

func (s *catalogService) UpdateProduct(ctx context.Context, in dto.UpdateProductInput) (dto.ProductView, error) {
	productID, err := parseID(in.ProductID, domain.ErrProductNotFound)
	if err != nil {
		return dto.ProductView{}, err
	}
	categoryID, err := parseID(in.CategoryID, domain.ErrCategoryNotFound)
	if err != nil {
		return dto.ProductView{}, err
	}

	product, err := s.productFactory.Update(ctx, domainsvc.UpdateProductInput{
		ProductID:   productID,
		Name:        in.Name,
		Description: in.Description,
		PricePoints: in.PricePoints,
		CategoryID:  categoryID,
		Sizes:       in.Sizes,
		PhotoKeys:   in.PhotoKeys,
		Active:      in.Active,
		Version:     in.Version,
	})
	if err != nil {
		return dto.ProductView{}, err
	}

	saved, err := s.products.Update(ctx, product, in.Version)
	if err != nil {
		return dto.ProductView{}, err
	}
	return dto.ToProductView(saved), nil
}

func (s *catalogService) DeactivateProduct(ctx context.Context, productID string) error {
	id, err := parseID(productID, domain.ErrProductNotFound)
	if err != nil {
		return err
	}
	return s.products.Deactivate(ctx, id)
}

func (s *catalogService) CreateCategory(ctx context.Context, in dto.CreateCategoryInput) (dto.CategoryView, error) {
	category, err := s.categoryFactory.Create(ctx, domainsvc.CreateCategoryInput{
		Code: in.Code,
		Name: in.Name,
	})
	if err != nil {
		return dto.CategoryView{}, err
	}

	saved, err := s.categories.Create(ctx, category)
	if err != nil {
		return dto.CategoryView{}, err
	}
	return dto.ToCategoryView(saved), nil
}

func (s *catalogService) UpdateCategory(ctx context.Context, in dto.UpdateCategoryInput) (dto.CategoryView, error) {
	id, err := parseID(in.ID, domain.ErrCategoryNotFound)
	if err != nil {
		return dto.CategoryView{}, err
	}

	category, err := s.categoryFactory.Update(ctx, domainsvc.UpdateCategoryInput{
		ID:     id,
		Name:   in.Name,
		Active: in.Active,
	})
	if err != nil {
		return dto.CategoryView{}, err
	}

	saved, err := s.categories.Update(ctx, category)
	if err != nil {
		return dto.CategoryView{}, err
	}
	return dto.ToCategoryView(saved), nil
}

func parseID(raw string, notFound error) (uuid.UUID, error) {
	id, err := uuid.Parse(raw)
	if err != nil {
		return uuid.Nil, notFound
	}
	return id, nil
}
