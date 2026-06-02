package handlers

import (
	"net/http"

	"github.com/mihnpro/Merch_shop/services/gateway/internal/domain"
	"github.com/mihnpro/Merch_shop/services/gateway/internal/infrastructure/rest/contract"
)

func (s *Server) Me(w http.ResponseWriter, r *http.Request) {
	c := domain.ClaimsFrom(r.Context())
	if c == nil {
		writeJSON(w, http.StatusUnauthorized, APIError{Code: "UNAUTHENTICATED", Message: "unauthenticated"})
		return
	}
	writeJSON(w, http.StatusOK, contract.MeResponse{
		UserID: c.UserID.String(),
		Role:   string(c.Role),
	})
}
