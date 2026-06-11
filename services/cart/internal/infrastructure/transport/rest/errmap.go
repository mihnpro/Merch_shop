package rest

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"

	"github.com/mihnpro/Merch_shop/services/cart/internal/domain"
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
	case errors.Is(err, domain.ErrCartNotFound),
		errors.Is(err, domain.ErrItemNotFound),
		errors.Is(err, domain.ErrProductNotFound):
		return http.StatusNotFound, apiError{Code: "NOT_FOUND", Message: err.Error()}
	case errors.Is(err, domain.ErrOutOfStock):
		return http.StatusConflict, apiError{Code: "OUT_OF_STOCK", Message: "out of stock"}
	case errors.Is(err, domain.ErrProductInactive):
		return http.StatusConflict, apiError{Code: "PRODUCT_INACTIVE", Message: "product is inactive"}
	case errors.Is(err, domain.ErrInvalidToken):
		return http.StatusUnauthorized, apiError{Code: "UNAUTHENTICATED", Message: "invalid token"}
	case errors.Is(err, domain.ErrForbidden):
		return http.StatusForbidden, apiError{Code: "PERMISSION_DENIED", Message: "permission denied"}
	case errors.Is(err, domain.ErrInvalidQuantity),
		errors.Is(err, domain.ErrInvalidSize),
		errors.Is(err, domain.ErrInvalidUserID),
		errors.Is(err, domain.ErrInvalidItemID),
		errors.Is(err, domain.ErrInvalidProductID),
		errors.Is(err, domain.ErrInvalidInput),
		errors.Is(err, domain.ErrEmptyRequest):
		return http.StatusBadRequest, apiError{Code: "INVALID_ARGUMENT", Message: err.Error()}
	default:
		fmt.Printf("[cart] unmapped error: %T %v\n", err, err)
		return http.StatusInternalServerError, apiError{Code: "INTERNAL", Message: "internal server error"}
	}
}
