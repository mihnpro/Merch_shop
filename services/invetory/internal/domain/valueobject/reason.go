package valueobject

import (
	"strings"
	"unicode/utf8"

	"github.com/mihnpro/Merch_shop/services/invetory/internal/domain"
)

const maxReasonLen = 500

type Reason struct {
	value string
}

func NewReason(raw string) (Reason, error) {
	v := strings.TrimSpace(raw)
	n := utf8.RuneCountInString(v)
	if n < 1 {
		return Reason{}, domain.ErrEmptyReason
	}
	if n > maxReasonLen {
		return Reason{}, domain.ErrReasonTooLong
	}
	return Reason{value: v}, nil
}

func NewReasonFromStored(v string) Reason {
	return Reason{value: v}
}

func (r Reason) String() string { return r.value }
