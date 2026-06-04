package account

import (
	"golang.org/x/crypto/bcrypt"

	vo "github.com/mihnpro/Merch_shop/services/user_customer/internal/domain/valueobject"
)

func (a *account) HashPassword(password vo.PlainPassword) (vo.PasswordHash, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password.Reveal()), bcrypt.DefaultCost)
	if err != nil {
		return vo.PasswordHash{}, err
	}
	return vo.NewPasswordHash(string(hash)), nil
}

func (a *account) VerifyPassword(password vo.PlainPassword, hash vo.PasswordHash) bool {
	return bcrypt.CompareHashAndPassword([]byte(hash.String()), []byte(password.Reveal())) == nil
}
