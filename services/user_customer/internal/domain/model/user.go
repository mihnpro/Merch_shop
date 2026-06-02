package model

import (
	"strings"
	"time"

	"github.com/google/uuid"

	"github.com/mihnpro/Merch_shop/services/user_customer/internal/domain"
	vo "github.com/mihnpro/Merch_shop/services/user_customer/internal/domain/valueobject"
)

type User struct {
	ID               uuid.UUID
	Login            vo.Login
	PasswordHash     vo.PasswordHash
	FirstName        string
	LastName         string
	Patronymic       string
	Email            vo.Email
	Phone            vo.PhoneNumber
	RoleID           uuid.UUID
	RoleCode         string
	Status           string
	FailedLoginCount int
	LockedUntil      *time.Time
	CreatedAt        time.Time
	UpdatedAt        time.Time
}

func NewUser(login vo.Login, hash vo.PasswordHash, firstName, lastName, patronymic string, email vo.Email, phone vo.PhoneNumber) (*User, error) {
	u := &User{
		Login:        login,
		PasswordHash: hash,
		FirstName:    strings.TrimSpace(firstName),
		LastName:     strings.TrimSpace(lastName),
		Patronymic:   strings.TrimSpace(patronymic),
		Email:        email,
		Phone:        phone,
	}
	if err := u.Validate(); err != nil {
		return nil, err
	}
	return u, nil
}

func (u *User) Validate() error {
	if u.Login.String() == "" {
		return domain.ErrEmptyLogin
	}
	if u.FirstName == "" {
		return domain.ErrEmptyFirstName
	}
	if u.LastName == "" {
		return domain.ErrEmptyLastName
	}
	if u.Email.String() == "" {
		return domain.ErrInvalidEmailFormat
	}
	return nil
}

func (u *User) IsLocked(now time.Time) bool {
	return u.LockedUntil != nil && now.Before(*u.LockedUntil)
}

func (u *User) Identity() Identity {
	return Identity{
		UserID: u.ID,
		Email:  u.Email.String(),
		Role:   u.RoleCode,
	}
}
