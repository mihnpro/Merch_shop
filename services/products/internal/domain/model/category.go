package model

import (
	"strings"
	"time"
	"unicode/utf8"

	"github.com/google/uuid"

	"github.com/mihnpro/Merch_shop/services/products/internal/domain"
	vo "github.com/mihnpro/Merch_shop/services/products/internal/domain/valueobject"
)

const maxCategoryNameLen = 100

type Category struct {
	ID        uuid.UUID
	Code      vo.CategoryCode
	Name      string
	Active    bool
	CreatedAt time.Time
	UpdatedAt time.Time
}

func NewCategory(code vo.CategoryCode, name string) (*Category, error) {
	c := &Category{
		Code:   code,
		Name:   strings.TrimSpace(name),
		Active: true,
	}
	if err := c.Validate(); err != nil {
		return nil, err
	}
	return c, nil
}

func (c *Category) Validate() error {
	if c.Code.String() == "" {
		return domain.ErrInvalidCategoryCode
	}
	n := utf8.RuneCountInString(c.Name)
	if n < 1 {
		return domain.ErrEmptyCategoryName
	}
	if n > maxCategoryNameLen {
		return domain.ErrCategoryNameTooLong
	}
	return nil
}

func (c *Category) Rename(name string) error {
	c.Name = strings.TrimSpace(name)
	return c.Validate()
}

func (c *Category) SetActive(active bool) {
	c.Active = active
}
