package valueobject

import (
	"unicode/utf8"

	"github.com/mihnpro/Merch_shop/services/user_customer/internal/domain"
)

const minPasswordLength = 8

type PlainPassword struct {
	value string
}

func NewPlainPassword(raw string) (PlainPassword, error) {
	if utf8.RuneCountInString(raw) < minPasswordLength {
		return PlainPassword{}, domain.ErrWeakPassword
	}
	return PlainPassword{value: raw}, nil
}

func PlainPasswordFromRaw(raw string) PlainPassword {
	return PlainPassword{value: raw}
}

func (p PlainPassword) Reveal() string { return p.value }

type PasswordHash struct {
	value string
}

func NewPasswordHash(hash string) PasswordHash {
	return PasswordHash{value: hash}
}

func NewPasswordHashFromStored(hash string) PasswordHash {
	return PasswordHash{value: hash}
}

func (h PasswordHash) String() string { return h.value }

func (h PasswordHash) IsZero() bool { return h.value == "" }
