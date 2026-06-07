package valueobject

import (
	"regexp"
	"strings"
	"unicode/utf8"

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

const maxSizeCodeLen = 20


type SizeCode struct {
	value string
}

func NewSizeCode(raw string) (SizeCode, error) {
	v := strings.TrimSpace(raw)
	n := utf8.RuneCountInString(v)
	if n < 1 || n > maxSizeCodeLen {
		return SizeCode{}, domain.ErrInvalidSizeCode
	}
	return SizeCode{value: v}, nil
}

func NewSizeCodeFromStored(v string) SizeCode {
	return SizeCode{value: v}
}

func (s SizeCode) String() string { return s.value }
