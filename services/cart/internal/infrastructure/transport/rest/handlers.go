package rest

import (
	"encoding/json"
	"net/http"

	"github.com/mihnpro/Merch_shop/services/cart/internal/app/dto"
	"github.com/mihnpro/Merch_shop/services/cart/internal/domain"
)

func (s *Server) GetCart(w http.ResponseWriter, r *http.Request) {
	identity, ok := IdentityFrom(r.Context())
	if !ok {
		writeError(w, domain.ErrInvalidToken)
		return
	}
	view, err := s.read.GetCart(r.Context(), dto.GetCartInput{UserID: identity.UserID.String()})
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, view)
}

type addItemRequest struct {
	ProductID string `json:"product_id"`
	Quantity  int    `json:"quantity"`
}

func (s *Server) AddItem(w http.ResponseWriter, r *http.Request) {
	identity, ok := IdentityFrom(r.Context())
	if !ok {
		writeError(w, domain.ErrInvalidToken)
		return
	}
	var req addItemRequest
	if !decodeJSON(w, r, &req) {
		return
	}
	item, err := s.write.AddItem(r.Context(), dto.AddItemInput{
		UserID:    identity.UserID.String(),
		ProductID: req.ProductID,
		Quantity:  req.Quantity,
	})
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, item)
}

type updateItemRequest struct {
	Quantity int `json:"quantity"`
}

func (s *Server) UpdateItem(w http.ResponseWriter, r *http.Request) {
	identity, ok := IdentityFrom(r.Context())
	if !ok {
		writeError(w, domain.ErrInvalidToken)
		return
	}
	itemID := r.PathValue("id")
	var req updateItemRequest
	if !decodeJSON(w, r, &req) {
		return
	}
	item, err := s.write.UpdateItem(r.Context(), dto.UpdateItemInput{
		UserID:      identity.UserID.String(),
		ItemID:      itemID,
		NewQuantity: req.Quantity,
	})
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, item)
}

func (s *Server) RemoveItem(w http.ResponseWriter, r *http.Request) {
	identity, ok := IdentityFrom(r.Context())
	if !ok {
		writeError(w, domain.ErrInvalidToken)
		return
	}
	itemID := r.PathValue("id")
	if err := s.write.RemoveItem(r.Context(), dto.RemoveItemInput{
		UserID: identity.UserID.String(),
		ItemID: itemID,
	}); err != nil {
		writeError(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (s *Server) ClearCart(w http.ResponseWriter, r *http.Request) {
	identity, ok := IdentityFrom(r.Context())
	if !ok {
		writeError(w, domain.ErrInvalidToken)
		return
	}
	if err := s.write.ClearCart(r.Context(), dto.ClearCartInput{
		UserID: identity.UserID.String(),
	}); err != nil {
		writeError(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func decodeJSON(w http.ResponseWriter, r *http.Request, dst any) bool {
	if err := json.NewDecoder(r.Body).Decode(dst); err != nil {
		writeJSON(w, http.StatusBadRequest, apiError{Code: "INVALID_JSON", Message: "invalid request body"})
		return false
	}
	return true
}
