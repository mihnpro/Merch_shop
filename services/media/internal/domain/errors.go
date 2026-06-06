package domain

import "errors"

var (
	ErrEmptyRequest = errors.New("empty request")

	ErrInvalidContentType = errors.New("invalid content type")

	ErrFileTooLarge = errors.New("file too large")

	ErrEmptyFile = errors.New("empty file")

	ErrStorageUnavailable = errors.New("storage unavailable")

	ErrUnauthenticated = errors.New("unauthenticated")

	ErrInvalidToken = errors.New("invalid token")

	ErrForbidden = errors.New("permission denied")
)
