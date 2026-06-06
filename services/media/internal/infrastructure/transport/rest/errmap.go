package rest

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/mihnpro/Merch_shop/services/media/internal/domain"
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
	case errors.Is(err, domain.ErrInvalidContentType):
		return http.StatusBadRequest, apiError{Code: "INVALID_ARGUMENT", Message: "unsupported content type: allowed image/jpeg, image/png, image/webp"}
	case errors.Is(err, domain.ErrFileTooLarge):
		return http.StatusBadRequest, apiError{Code: "INVALID_ARGUMENT", Message: "file too large: max 5 MB"}
	case errors.Is(err, domain.ErrEmptyFile):
		return http.StatusBadRequest, apiError{Code: "INVALID_ARGUMENT", Message: "empty file"}
	case errors.Is(err, domain.ErrEmptyRequest):
		return http.StatusBadRequest, apiError{Code: "INVALID_ARGUMENT", Message: "expected multipart form with a \"file\" field"}
	case errors.Is(err, domain.ErrUnauthenticated), errors.Is(err, domain.ErrInvalidToken):
		return http.StatusUnauthorized, apiError{Code: "UNAUTHENTICATED", Message: "unauthenticated"}
	case errors.Is(err, domain.ErrForbidden):
		return http.StatusForbidden, apiError{Code: "PERMISSION_DENIED", Message: "admin role required"}
	case errors.Is(err, domain.ErrStorageUnavailable):
		return http.StatusServiceUnavailable, apiError{Code: "UNAVAILABLE", Message: "storage unavailable"}
	default:
		return http.StatusInternalServerError, apiError{Code: "INTERNAL", Message: "internal server error"}
	}
}
