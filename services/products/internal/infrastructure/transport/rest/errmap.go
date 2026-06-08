package rest

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/mihnpro/Merch_shop/services/products/internal/domain"
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
	case errors.Is(err, domain.ErrProductNotFound):
		return http.StatusNotFound, apiError{Code: "NOT_FOUND", Message: "product not found"}
	case errors.Is(err, domain.ErrCategoryNotFound):
		return http.StatusNotFound, apiError{Code: "NOT_FOUND", Message: "category not found"}
	case errors.Is(err, domain.ErrVersionConflict):
		return http.StatusConflict, apiError{Code: "VERSION_CONFLICT", Message: "product was modified by another request"}
	case errors.Is(err, domain.ErrCategoryAlreadyExists):
		return http.StatusConflict, apiError{Code: "ALREADY_EXISTS", Message: "category with this code already exists"}
	case errors.Is(err, domain.ErrCategoryNotActive):
		return http.StatusConflict, apiError{Code: "CATEGORY_NOT_ACTIVE", Message: "category is not active"}
	case errors.Is(err, domain.ErrInvalidToken):
		return http.StatusUnauthorized, apiError{Code: "UNAUTHENTICATED", Message: "invalid token"}
	case errors.Is(err, domain.ErrForbidden):
		return http.StatusForbidden, apiError{Code: "PERMISSION_DENIED", Message: "permission denied"}
	case errors.Is(err, domain.ErrInvalidInput),
		errors.Is(err, domain.ErrEmptyRequest),
		errors.Is(err, domain.ErrInvalidPrice),
		errors.Is(err, domain.ErrInvalidPhotoKey),
		errors.Is(err, domain.ErrInvalidCategoryCode),
		errors.Is(err, domain.ErrEmptyName),
		errors.Is(err, domain.ErrNameTooLong),
		errors.Is(err, domain.ErrEmptyDescription),
		errors.Is(err, domain.ErrDescriptionTooLong),
		errors.Is(err, domain.ErrEmptyCategoryName),
		errors.Is(err, domain.ErrCategoryNameTooLong):
		return http.StatusBadRequest, apiError{Code: "INVALID_ARGUMENT", Message: err.Error()}
	default:
		return http.StatusInternalServerError, apiError{Code: "INTERNAL", Message: "internal server error"}
	}
}
