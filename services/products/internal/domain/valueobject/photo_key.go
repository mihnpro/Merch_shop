package valueobject

import (
	"regexp"
	"strings"

	"github.com/mihnpro/Merch_shop/services/products/internal/domain"
)

var photoKeyRegex = regexp.MustCompile(`^products/[0-9a-fA-F-]{36}\.(jpg|jpeg|png|webp)$`)

type PhotoKey struct {
	value string
}

func NewPhotoKey(raw string) (PhotoKey, error) {
	v := strings.TrimSpace(raw)
	if v == "" {
		return PhotoKey{}, nil
	}
	if !photoKeyRegex.MatchString(v) {
		return PhotoKey{}, domain.ErrInvalidPhotoKey
	}
	return PhotoKey{value: v}, nil
}

func NewPhotoKeyFromStored(v string) PhotoKey {
	return PhotoKey{value: v}
}

func (k PhotoKey) String() string { return k.value }

func (k PhotoKey) IsEmpty() bool { return k.value == "" }
