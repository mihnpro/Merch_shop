package valueobject

import (
	"regexp"
	"strings"

	"github.com/mihnpro/Merch_shop/services/user_customer/internal/domain"
)

var phoneRegex = regexp.MustCompile(`^\+?[0-9]{10,15}$`)

type PhoneNumber struct {
	value string
}

func NewPhoneNumber(raw string) (PhoneNumber, error) {
	v := strings.TrimSpace(raw)
	if v == "" {
		return PhoneNumber{}, nil
	}
	if !phoneRegex.MatchString(v) {
		return PhoneNumber{}, domain.ErrInvalidPhoneFormat
	}
	return PhoneNumber{value: v}, nil
}

func NewPhoneNumberFromStored(v string) PhoneNumber {
	return PhoneNumber{value: v}
}

func (p PhoneNumber) String() string { return p.value }

func (p PhoneNumber) IsEmpty() bool { return p.value == "" }
