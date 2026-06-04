package model

import (
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/mihnpro/Merch_shop/services/user_customer/internal/domain"
	vo "github.com/mihnpro/Merch_shop/services/user_customer/internal/domain/valueobject"
)

func mustLogin(t *testing.T, raw string) vo.Login {
	t.Helper()
	l, err := vo.NewLogin(raw)
	if err != nil {
		t.Fatalf("NewLogin(%q): %v", raw, err)
	}
	return l
}

func mustEmail(t *testing.T, raw string) vo.Email {
	t.Helper()
	e, err := vo.NewEmail(raw)
	if err != nil {
		t.Fatalf("NewEmail(%q): %v", raw, err)
	}
	return e
}

func mustPhone(t *testing.T, raw string) vo.PhoneNumber {
	t.Helper()
	p, err := vo.NewPhoneNumber(raw)
	if err != nil {
		t.Fatalf("NewPhoneNumber(%q): %v", raw, err)
	}
	return p
}

func TestUser_Validate(t *testing.T) {
	valid := func() *User {
		return &User{
			Login:        mustLogin(t, "testuser"),
			PasswordHash: vo.NewPasswordHash("hash"),
			FirstName:    "John",
			LastName:     "Doe",
			Email:        mustEmail(t, "john@example.com"),
		}
	}

	tests := []struct {
		name    string
		mutate  func(u *User)
		wantErr error
	}{
		{"valid user", func(*User) {}, nil},
		{"empty login", func(u *User) { u.Login = vo.Login{} }, domain.ErrEmptyLogin},
		{"empty first name", func(u *User) { u.FirstName = "" }, domain.ErrEmptyFirstName},
		{"empty last name", func(u *User) { u.LastName = "" }, domain.ErrEmptyLastName},
		{"empty email", func(u *User) { u.Email = vo.Email{} }, domain.ErrInvalidEmailFormat},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			u := valid()
			tt.mutate(u)
			if err := u.Validate(); !errors.Is(err, tt.wantErr) {
				t.Errorf("Validate() = %v, want %v", err, tt.wantErr)
			}
		})
	}
}

func TestNewUser(t *testing.T) {
	t.Run("trims names and validates", func(t *testing.T) {
		u, err := NewUser(
			mustLogin(t, "testuser"), vo.NewPasswordHash("hash"),
			"  John  ", "  Doe  ", "  Smith  ",
			mustEmail(t, "john@example.com"), mustPhone(t, "+1234567890"),
		)
		if err != nil {
			t.Fatalf("NewUser() unexpected error: %v", err)
		}
		if u.FirstName != "John" || u.LastName != "Doe" || u.Patronymic != "Smith" {
			t.Errorf("names not trimmed: %q %q %q", u.FirstName, u.LastName, u.Patronymic)
		}
	})

	t.Run("invalid user returns error and nil", func(t *testing.T) {
		u, err := NewUser(
			vo.Login{}, vo.NewPasswordHash("hash"),
			"John", "Doe", "",
			mustEmail(t, "john@example.com"), vo.PhoneNumber{},
		)
		if !errors.Is(err, domain.ErrEmptyLogin) {
			t.Errorf("err = %v, want %v", err, domain.ErrEmptyLogin)
		}
		if u != nil {
			t.Errorf("expected nil user on error, got %+v", u)
		}
	})
}

func TestUser_IsLocked(t *testing.T) {
	now := time.Date(2026, 6, 3, 12, 0, 0, 0, time.UTC)
	past, future := now.Add(-time.Hour), now.Add(time.Hour)

	tests := []struct {
		name        string
		lockedUntil *time.Time
		want        bool
	}{
		{"never locked", nil, false},
		{"lock in the past", &past, false},
		{"lock in the future", &future, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			u := &User{LockedUntil: tt.lockedUntil}
			if got := u.IsLocked(now); got != tt.want {
				t.Errorf("IsLocked() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestUser_Identity(t *testing.T) {
	id := uuid.New()
	u := &User{ID: id, Email: mustEmail(t, "john@example.com"), Role: "admin"}

	got := u.Identity()
	want := Identity{UserID: id, Email: "john@example.com", Role: "admin"}
	if got != want {
		t.Errorf("Identity() = %+v, want %+v", got, want)
	}
}
