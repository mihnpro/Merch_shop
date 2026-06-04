package valueobject

import (
	"regexp"
	"strings"

	"github.com/mihnpro/Merch_shop/services/user_customer/internal/domain"
)

var emailRegex = regexp.MustCompile(`^[a-zA-Z0-9._%+\-]+@[a-zA-Z0-9.\-]+\.[a-zA-Z]{2,}$`)

type Email struct {
	value string
}

func NewEmail(raw string) (Email, error) {
	v := strings.TrimSpace(raw)
	if !emailRegex.MatchString(v) {
		return Email{}, domain.ErrInvalidEmailFormat
	}
	return Email{value: v}, nil
}

func NewEmailFromStored(v string) Email {
	return Email{value: v}
}

func (e Email) String() string { return e.value }
