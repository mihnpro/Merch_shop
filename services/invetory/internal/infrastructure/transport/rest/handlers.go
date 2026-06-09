package rest

import (
	"encoding/json"
	"net/http"

	"github.com/mihnpro/Merch_shop/services/invetory/internal/app/dto"
)

type adjustStockRequest struct {
	ProductID   string `json:"product_id"`
	Delta       int    `json:"delta"`
	OperationID string `json:"operation_id"`
	Reason      string `json:"reason"`
}

type adjustStockResponse struct {
	ProductID string `json:"product_id"`
	Available int    `json:"available"`
	Version   int    `json:"version"`
}

type stockItemResponse struct {
	ProductID string `json:"product_id"`
	Available int    `json:"available"`
	Version   int    `json:"version"`
}

type listStockResponse struct {
	Items []stockItemResponse `json:"items"`
}

func (s *Server) ListStock(w http.ResponseWriter, r *http.Request) {
	views, err := s.read.ListStock(r.Context())
	if err != nil {
		writeError(w, err)
		return
	}
	items := make([]stockItemResponse, 0, len(views))
	for _, v := range views {
		items = append(items, stockItemResponse{
			ProductID: v.ProductID,
			Available: v.Available,
			Version:   v.Version,
		})
	}
	writeJSON(w, http.StatusOK, listStockResponse{Items: items})
}

func (s *Server) AdjustStock(w http.ResponseWriter, r *http.Request) {
	var req adjustStockRequest
	if !decodeJSON(w, r, &req) {
		return
	}
	view, err := s.write.AdjustStock(r.Context(), dto.AdjustStockInput{
		ProductID:   req.ProductID,
		Delta:       req.Delta,
		OperationID: req.OperationID,
		Reason:      req.Reason,
	})
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, adjustStockResponse{
		ProductID: view.ProductID,
		Available: view.Available,
		Version:   view.Version,
	})
}

func decodeJSON(w http.ResponseWriter, r *http.Request, dst any) bool {
	if err := json.NewDecoder(r.Body).Decode(dst); err != nil {
		writeJSON(w, http.StatusBadRequest, apiError{Code: "INVALID_JSON", Message: "invalid request body"})
		return false
	}
	return true
}
