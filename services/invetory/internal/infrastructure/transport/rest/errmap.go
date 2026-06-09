package rest

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/mihnpro/Merch_shop/services/invetory/internal/domain"
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
	case errors.Is(err, domain.ErrStockNotFound):
		return http.StatusNotFound, apiError{Code: "NOT_FOUND", Message: "stock not found"}
	case errors.Is(err, domain.ErrReservationNotFound):
		return http.StatusNotFound, apiError{Code: "NOT_FOUND", Message: "reservation not found"}
	case errors.Is(err, domain.ErrInsufficientStock):
		return http.StatusConflict, apiError{Code: "INSUFFICIENT_STOCK", Message: "insufficient stock"}
	case errors.Is(err, domain.ErrVersionConflict):
		return http.StatusConflict, apiError{Code: "VERSION_CONFLICT", Message: "stock was modified by another request"}
	case errors.Is(err, domain.ErrInvalidToken):
		return http.StatusUnauthorized, apiError{Code: "UNAUTHENTICATED", Message: "invalid token"}
	case errors.Is(err, domain.ErrForbidden):
		return http.StatusForbidden, apiError{Code: "PERMISSION_DENIED", Message: "permission denied"}
	case errors.Is(err, domain.ErrInvalidProductID),
		errors.Is(err, domain.ErrInvalidOrderID),
		errors.Is(err, domain.ErrInvalidOperationID),
		errors.Is(err, domain.ErrInvalidQuantity),
		errors.Is(err, domain.ErrZeroDelta),
		errors.Is(err, domain.ErrEmptyReason),
		errors.Is(err, domain.ErrReasonTooLong),
		errors.Is(err, domain.ErrEmptyReservation),
		errors.Is(err, domain.ErrInvalidInput),
		errors.Is(err, domain.ErrEmptyRequest):
		return http.StatusBadRequest, apiError{Code: "INVALID_ARGUMENT", Message: err.Error()}
	default:
		return http.StatusInternalServerError, apiError{Code: "INTERNAL", Message: "internal server error"}
	}
}
