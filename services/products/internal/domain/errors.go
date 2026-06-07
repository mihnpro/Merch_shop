package domain

import "errors"

var (
	ErrInvalidInput = errors.New("invalid input")
	ErrEmptyRequest = errors.New("empty request")
	ErrInvalidToken = errors.New("invalid token")
	ErrForbidden    = errors.New("permission denied")

	ErrProductNotFound    = errors.New("product not found")
	ErrEmptyName          = errors.New("product name cannot be empty")
	ErrNameTooLong        = errors.New("product name is too long")
	ErrEmptyDescription   = errors.New("product description cannot be empty")
	ErrDescriptionTooLong = errors.New("product description is too long")
	ErrInvalidPrice       = errors.New("price must be a positive number of points")
	ErrInvalidPhotoKey    = errors.New("invalid photo key format")
	ErrVersionConflict    = errors.New("product was modified by another request")

	ErrCategoryNotFound      = errors.New("category not found")
	ErrCategoryNotActive     = errors.New("category is not active")
	ErrCategoryAlreadyExists = errors.New("category with this code already exists")
	ErrInvalidCategoryCode   = errors.New("invalid category code")
	ErrEmptyCategoryName     = errors.New("category name cannot be empty")
	ErrCategoryNameTooLong   = errors.New("category name is too long")

	ErrInvalidSizeCode = errors.New("invalid size code")
)
