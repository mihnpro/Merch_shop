package valueobject

import (
	"strings"

	"github.com/mihnpro/Merch_shop/services/user_customer/internal/domain"
)

type Login struct {
	value string
}

func NewLogin(raw string) (Login, error) {
	v := strings.TrimSpace(raw)
	if v == "" {
		return Login{}, domain.ErrEmptyLogin
	}
	return Login{value: v}, nil
}

func NewLoginFromStored(v string) Login {
	return Login{value: v}
}

func (l Login) String() string { return l.value }
