package rest

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"

	"github.com/mihnpro/Merch_shop/services/order/internal/domain"
)

type apiError struct {
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

func mapError(err error) (int, apiError) {
	switch {
	case errors.Is(err, domain.ErrOrderNotFound):
		return http.StatusNotFound, apiError{Code: "NOT_FOUND", Message: err.Error()}
	case errors.Is(err, domain.ErrInvalidToken):
		return http.StatusUnauthorized, apiError{Code: "UNAUTHENTICATED", Message: "invalid token"}
	case errors.Is(err, domain.ErrForbidden):
		return http.StatusForbidden, apiError{Code: "PERMISSION_DENIED", Message: "permission denied"}
	case errors.Is(err, domain.ErrEmptyCart):
		return http.StatusConflict, apiError{Code: "EMPTY_CART", Message: err.Error()}
	case errors.Is(err, domain.ErrInsufficientBalance):
		return http.StatusPaymentRequired, apiError{Code: "INSUFFICIENT_BALANCE", Message: err.Error()}
	case errors.Is(err, domain.ErrInvalidStatusChange):
		return http.StatusConflict, apiError{Code: "INVALID_STATUS_CHANGE", Message: err.Error()}
	case errors.Is(err, domain.ErrInvalidInput),
		errors.Is(err, domain.ErrEmptyRequest):
		return http.StatusBadRequest, apiError{Code: "INVALID_ARGUMENT", Message: err.Error()}
	default:
		fmt.Printf("[order] unmapped error: %T %v\n", err, err)
		return http.StatusInternalServerError, apiError{Code: "INTERNAL", Message: "internal server error"}
	}
}
