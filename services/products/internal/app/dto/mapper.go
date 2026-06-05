package dto

import "github.com/mihnpro/Merch_shop/services/products/internal/domain/model"

func ToCategoryView(c *model.Category) CategoryView {
	return CategoryView{
		ID:        c.ID.String(),
		Code:      c.Code.String(),
		Name:      c.Name,
		Active:    c.Active,
		CreatedAt: c.CreatedAt,
		UpdatedAt: c.UpdatedAt,
	}
}

func ToProductView(p *model.Product) ProductView {
	category := p.Category
	return ProductView{
		ID:          p.ID.String(),
		Name:        p.Name,
		Description: p.Description,
		PricePoints: p.Price.Int64(),
		Category:    ToCategoryView(&category),
		Sizes:       p.SizeCodes(),
		PhotoKey:    p.PhotoKey.String(),
		Active:      p.Active,
		Version:     p.Version,
		CreatedAt:   p.CreatedAt,
		UpdatedAt:   p.UpdatedAt,
	}
}
