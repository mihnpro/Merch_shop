package handlers

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/mihnpro/Merch_shop/services/gateway/internal/domain"
)

type APIError struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

func writeError(w http.ResponseWriter, err error) {
	status, body := mapError(err)
	writeJSON(w, status, body)
}

func mapError(err error) (int, APIError) {
	switch {
	case errors.Is(err, domain.ErrBadRequest):
		return http.StatusBadRequest, APIError{Code: "BAD_REQUEST", Message: "bad request"}
	case errors.Is(err, domain.ErrUnauthenticated):
		return http.StatusUnauthorized, APIError{Code: "UNAUTHENTICATED", Message: "unauthenticated"}
	case errors.Is(err, domain.ErrForbidden):
		return http.StatusForbidden, APIError{Code: "PERMISSION_DENIED", Message: "forbidden"}
	case errors.Is(err, domain.ErrNotFound):
		return http.StatusNotFound, APIError{Code: "NOT_FOUND", Message: "not found"}
	case errors.Is(err, domain.ErrAlreadyExists):
		return http.StatusConflict, APIError{Code: "ALREADY_EXISTS", Message: "already exists"}
	case errors.Is(err, domain.ErrFailedPrecondition):
		return http.StatusUnprocessableEntity, APIError{Code: "FAILED_PRECONDITION", Message: "failed precondition"}
	case errors.Is(err, domain.ErrRateLimited):
		return http.StatusTooManyRequests, APIError{Code: "RATE_LIMITED", Message: "rate limited"}
	case errors.Is(err, domain.ErrUnavailable):
		return http.StatusServiceUnavailable, APIError{Code: "UNAVAILABLE", Message: "service unavailable"}
	case errors.Is(err, domain.ErrTimeout):
		return http.StatusGatewayTimeout, APIError{Code: "TIMEOUT", Message: "timeout"}
	default:
		return http.StatusInternalServerError, APIError{Code: "INTERNAL", Message: "internal error"}
	}
}
