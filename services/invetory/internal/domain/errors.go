package domain

import "errors"

var (
	ErrInvalidInput = errors.New("invalid input")
	ErrEmptyRequest = errors.New("empty request")
	ErrInvalidToken = errors.New("invalid token")
	ErrForbidden    = errors.New("permission denied")

	// Остатки.
	ErrStockNotFound     = errors.New("stock not found")
	ErrInsufficientStock = errors.New("insufficient stock")
	ErrVersionConflict   = errors.New("stock was modified by another request")
	ErrInvalidProductID  = errors.New("invalid product id")

	// Корректировка остатка.
	ErrZeroDelta          = errors.New("adjustment delta must be non-zero")
	ErrEmptyReason        = errors.New("reason cannot be empty")
	ErrReasonTooLong      = errors.New("reason is too long")
	ErrInvalidOperationID = errors.New("invalid operation id")

	// Резервы.
	ErrReservationNotFound = errors.New("reservation not found")
	ErrEmptyReservation    = errors.New("reservation must contain at least one item")
	ErrInvalidQuantity     = errors.New("quantity must be between 1 and 100")
	ErrInvalidOrderID      = errors.New("invalid order id")
	ErrAlreadyReleased     = errors.New("reservation already released")
)
