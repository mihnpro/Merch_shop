package domain

import "errors"

var (
	ErrUserAlreadyExists  = errors.New("user already exists")
	ErrInvalidCredentials = errors.New("invalid credentials")
	ErrUserNotFound       = errors.New("user not found")
	ErrInvalidInput       = errors.New("invalid input")
	ErrEmptyRequest       = errors.New("empty request")
	ErrEmptyLogin         = errors.New("login cannot be empty")
	ErrWeakPassword       = errors.New("password must be at least 8 characters")
	ErrEmptyFirstName     = errors.New("first name cannot be empty")
	ErrEmptyLastName      = errors.New("last name cannot be empty")
	ErrInvalidEmailFormat = errors.New("invalid email format")
	ErrInvalidPhoneFormat = errors.New("invalid phone number format")
	ErrInvalidToken       = errors.New("invalid token")
	ErrForbidden          = errors.New("permission denied")
)
