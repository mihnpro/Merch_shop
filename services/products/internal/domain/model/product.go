package model

import (
	"strings"
	"time"
	"unicode/utf8"

	"github.com/google/uuid"

	"github.com/mihnpro/Merch_shop/services/products/internal/domain"
	vo "github.com/mihnpro/Merch_shop/services/products/internal/domain/valueobject"
)

const (
	maxNameLen        = 500
	maxDescriptionLen = 5000
)

type Product struct {
	ID          uuid.UUID
	Name        string
	Description string
	Price       vo.PricePoints
	Category    Category
	PhotoKeys   []vo.PhotoKey
	Active      bool
	Version     int
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

func NewProduct(name, description string, price vo.PricePoints, category Category, photoKeys []vo.PhotoKey) (*Product, error) {
	p := &Product{
		Name:        strings.TrimSpace(name),
		Description: strings.TrimSpace(description),
		Price:       price,
		Category:    category,
		PhotoKeys:   photoKeys,
		Active:      true,
		Version:     1,
	}
	if err := p.Validate(); err != nil {
		return nil, err
	}
	return p, nil
}

func (p *Product) Validate() error {
	n := utf8.RuneCountInString(p.Name)
	if n < 1 {
		return domain.ErrEmptyName
	}
	if n > maxNameLen {
		return domain.ErrNameTooLong
	}
	d := utf8.RuneCountInString(p.Description)
	if d < 1 {
		return domain.ErrEmptyDescription
	}
	if d > maxDescriptionLen {
		return domain.ErrDescriptionTooLong
	}
	if p.Price.Int64() <= 0 {
		return domain.ErrInvalidPrice
	}
	if p.Category.ID == uuid.Nil {
		return domain.ErrCategoryNotFound
	}
	if !p.Category.Active {
		return domain.ErrCategoryNotActive
	}
	return nil
}

func (p *Product) ApplyUpdate(name, description string, price vo.PricePoints, category Category, photoKeys []vo.PhotoKey, active bool, expectedVersion int) error {
	if p.Version != expectedVersion {
		return domain.ErrVersionConflict
	}
	p.Name = strings.TrimSpace(name)
	p.Description = strings.TrimSpace(description)
	p.Price = price
	p.Category = category
	p.PhotoKeys = photoKeys
	p.Active = active
	return p.Validate()
}


func (p *Product) Deactivate() {
	p.Active = false
}

func (p *Product) PhotoKeyStrings() []string {
	keys := make([]string, 0, len(p.PhotoKeys))
	for _, k := range p.PhotoKeys {
		keys = append(keys, k.String())
	}
	return keys
}
