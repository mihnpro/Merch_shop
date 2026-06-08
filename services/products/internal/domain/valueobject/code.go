package valueobject

import (
	"regexp"
	"strings"

	"github.com/mihnpro/Merch_shop/services/products/internal/domain"
)

var categoryCodeRegex = regexp.MustCompile(`^[a-z_]{1,30}$`)

type CategoryCode struct {
	value string
}

func NewCategoryCode(raw string) (CategoryCode, error) {
	v := strings.TrimSpace(raw)
	if !categoryCodeRegex.MatchString(v) {
		return CategoryCode{}, domain.ErrInvalidCategoryCode
	}
	return CategoryCode{value: v}, nil
}

func NewCategoryCodeFromStored(v string) CategoryCode {
	return CategoryCode{value: v}
}

func (c CategoryCode) String() string { return c.value }
