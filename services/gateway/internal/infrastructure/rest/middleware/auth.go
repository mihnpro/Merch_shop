package middleware

import (
	"encoding/json"
	"net/http"
	"strings"
)

const bearerPrefix = "Bearer "

func (m *Middleware) Authorize(w http.ResponseWriter, r *http.Request) bool {
	token, ok := bearerToken(r)
	if !ok {
		unauthorized(w)
		return false
	}

	if err := m.verify.VerifyAccess(token); err != nil {
		unauthorized(w)
		return false
	}

	return true
}

func bearerToken(r *http.Request) (string, bool) {
	h := r.Header.Get("Authorization")
	if !strings.HasPrefix(h, bearerPrefix) {
		return "", false
	}
	token := strings.TrimSpace(strings.TrimPrefix(h, bearerPrefix))
	return token, token != ""
}

func unauthorized(w http.ResponseWriter) {
	writeError(w, http.StatusUnauthorized, "UNAUTHENTICATED", "unauthenticated")
}

func writeError(w http.ResponseWriter, status int, code, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(map[string]string{
		"code":    code,
		"message": message,
	})
}
