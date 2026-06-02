package domain

import "errors"

var (
	ErrBadRequest         = errors.New("bad request")
	ErrUnauthenticated    = errors.New("unauthenticated")
	ErrForbidden          = errors.New("forbidden")
	ErrNotFound           = errors.New("not found")
	ErrAlreadyExists      = errors.New("already exists")
	ErrFailedPrecondition = errors.New("failed precondition")
	ErrRateLimited        = errors.New("rate limited")
	ErrUnavailable        = errors.New("service unavailable")
	ErrTimeout            = errors.New("timeout")
	ErrInternal           = errors.New("internal error")
)
