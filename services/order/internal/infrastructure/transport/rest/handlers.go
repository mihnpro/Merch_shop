package rest

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/mihnpro/Merch_shop/services/order/internal/app/dto"
	"github.com/mihnpro/Merch_shop/services/order/internal/domain"
)

func (s *Server) CreateOrder(w http.ResponseWriter, r *http.Request) {
	identity, ok := IdentityFrom(r.Context())
	if !ok {
		writeError(w, domain.ErrInvalidToken)
		return
	}
	var body struct {
		DeliveryAddress string  `json:"delivery_address"`
		Note            *string `json:"note"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil || body.DeliveryAddress == "" {
		writeError(w, domain.ErrInvalidInput)
		return
	}
	if body.Note != nil && len(*body.Note) > 1000 {
		writeError(w, domain.ErrInvalidInput)
		return
	}
	view, err := s.orders.CreateOrder(r.Context(), dto.CreateOrderInput{
		UserID:          identity.UserID.String(),
		DeliveryAddress: body.DeliveryAddress,
		Note:            body.Note,
	})
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, view)
}

func (s *Server) GetMyOrders(w http.ResponseWriter, r *http.Request) {
	identity, ok := IdentityFrom(r.Context())
	if !ok {
		writeError(w, domain.ErrInvalidToken)
		return
	}
	q := r.URL.Query()
	pageSize := 20
	if v := q.Get("page_size"); v != "" {
		var n int
		if _, err := fmt.Sscanf(v, "%d", &n); err == nil && n > 0 {
			pageSize = n
		}
	}
	views, nextToken, err := s.query.ListOrders(r.Context(), dto.ListOrdersInput{
		UserID:    identity.UserID.String(),
		Status:    q.Get("status"),
		PageSize:  pageSize,
		PageToken: q.Get("page_token"),
	})
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"orders": views, "next_page_token": nextToken})
}

func (s *Server) GetOrder(w http.ResponseWriter, r *http.Request) {
	identity, ok := IdentityFrom(r.Context())
	if !ok {
		writeError(w, domain.ErrInvalidToken)
		return
	}
	id := r.PathValue("id")
	view, err := s.query.GetOrder(r.Context(), dto.GetOrderInput{
		OrderID: id,
	})
	if err != nil {
		writeError(w, err)
		return
	}
	if identity.Role != "admin" && view.UserID != identity.UserID.String() {
		writeError(w, domain.ErrForbidden)
		return
	}
	writeJSON(w, http.StatusOK, view)
}

func (s *Server) AdminUpdateOrderStatus(w http.ResponseWriter, r *http.Request) {
	if _, ok := IdentityFrom(r.Context()); !ok {
		writeError(w, domain.ErrInvalidToken)
		return
	}
	id := r.PathValue("id")
	var body struct {
		Status string `json:"status"`
		Reason string `json:"reason"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil || body.Status == "" {
		writeError(w, domain.ErrInvalidInput)
		return
	}
	view, err := s.orders.UpdateOrderStatus(r.Context(), dto.UpdateOrderStatusInput{
		OrderID:   id,
		NewStatus: body.Status,
		Reason:    body.Reason,
	})
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, view)
}

func (s *Server) AdminAnalytics(w http.ResponseWriter, r *http.Request) {
	period := r.URL.Query().Get("period")
	view, err := s.query.GetAnalytics(r.Context(), period)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, view)
}

func (s *Server) AdminListOrders(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	views, nextToken, err := s.query.ListOrders(r.Context(), dto.ListOrdersInput{
		UserID:    q.Get("user_id"),
		Status:    q.Get("status"),
		PageSize:  20,
		PageToken: q.Get("page_token"),
	})
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"orders": views, "next_page_token": nextToken})
}
