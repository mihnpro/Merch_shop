package domain

import "errors"

var (
	ErrOrderNotFound        = errors.New("order not found")
	ErrEmptyCart            = errors.New("cart is empty")
	ErrInsufficientBalance  = errors.New("insufficient points balance")
	ErrInvalidStatusChange  = errors.New("invalid order status transition")
	ErrInvalidInput         = errors.New("invalid input")
	ErrEmptyRequest         = errors.New("empty request")
	ErrInvalidToken         = errors.New("invalid token")
	ErrForbidden            = errors.New("permission denied")
)
