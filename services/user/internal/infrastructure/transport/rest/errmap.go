package rest

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/mihnpro/Merch_shop/services/user_customer/internal/domain"
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
	case errors.Is(err, domain.ErrUserAlreadyExists):
		return http.StatusConflict, apiError{Code: "ALREADY_EXISTS", Message: "user already exists"}
	case errors.Is(err, domain.ErrInvalidCredentials):
		return http.StatusUnauthorized, apiError{Code: "UNAUTHENTICATED", Message: "invalid credentials"}
	case errors.Is(err, domain.ErrInvalidToken):
		return http.StatusUnauthorized, apiError{Code: "UNAUTHENTICATED", Message: "invalid token"}
	case errors.Is(err, domain.ErrForbidden):
		return http.StatusForbidden, apiError{Code: "PERMISSION_DENIED", Message: "permission denied"}
	case errors.Is(err, domain.ErrUserNotFound):
		return http.StatusNotFound, apiError{Code: "NOT_FOUND", Message: "user not found"}
	case errors.Is(err, domain.ErrInvalidInput),
		errors.Is(err, domain.ErrEmptyRequest),
		errors.Is(err, domain.ErrEmptyLogin),
		errors.Is(err, domain.ErrWeakPassword),
		errors.Is(err, domain.ErrEmptyFirstName),
		errors.Is(err, domain.ErrEmptyLastName),
		errors.Is(err, domain.ErrInvalidEmailFormat),
		errors.Is(err, domain.ErrInvalidPhoneFormat),
		errors.Is(err, domain.ErrInvalidRole),
		errors.Is(err, domain.ErrInvalidStatus):
		return http.StatusBadRequest, apiError{Code: "INVALID_ARGUMENT", Message: err.Error()}
	case errors.Is(err, domain.ErrInsufficientBalance):
		return http.StatusUnprocessableEntity, apiError{Code: "INSUFFICIENT_BALANCE", Message: "insufficient points balance"}
	case errors.Is(err, domain.ErrOperationAlreadyDone):
		return http.StatusConflict, apiError{Code: "ALREADY_PROCESSED", Message: "operation already processed"}
	case errors.Is(err, domain.ErrUserBlocked):
		return http.StatusForbidden, apiError{Code: "USER_BLOCKED", Message: "user is blocked"}
	default:
		return http.StatusInternalServerError, apiError{Code: "INTERNAL", Message: "internal server error"}
	}
}
