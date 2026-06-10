package domain

import "errors"

var (
	ErrInvalidInput  = errors.New("invalid input")
	ErrEmptyRequest  = errors.New("empty request")
	ErrInvalidToken  = errors.New("invalid token")
	ErrForbidden     = errors.New("permission denied")

	ErrCartNotFound    = errors.New("cart not found")
	ErrItemNotFound    = errors.New("cart item not found")
	ErrOutOfStock      = errors.New("out of stock")
	ErrProductInactive = errors.New("product is inactive")
	ErrProductNotFound = errors.New("product not found")
	ErrInvalidQuantity = errors.New("quantity must be between 1 and 9999")
	ErrInvalidSize     = errors.New("invalid size: must be one of XS, S, M, L, XL, XXL")
	ErrInvalidUserID   = errors.New("invalid user id")
	ErrInvalidItemID   = errors.New("invalid item id")
	ErrInvalidProductID = errors.New("invalid product id")
)
