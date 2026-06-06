package rest

import (
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"github.com/mihnpro/Merch_shop/services/products/internal/app/dto"
)

type categoryResponse struct {
	ID        string    `json:"id"`
	Code      string    `json:"code"`
	Name      string    `json:"name"`
	Active    bool      `json:"active"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type productResponse struct {
	ID          string           `json:"id"`
	Name        string           `json:"name"`
	Description string           `json:"description"`
	PricePoints int64            `json:"price_points"`
	Category    categoryResponse `json:"category"`
	Sizes       []string         `json:"sizes"`
	PhotoKeys   []string         `json:"photo_keys"`
	Active      bool             `json:"active"`
	Version     int              `json:"version"`
	CreatedAt   time.Time        `json:"created_at"`
	UpdatedAt   time.Time        `json:"updated_at"`
}

type listProductsResponse struct {
	Products      []productResponse `json:"products"`
	NextPageToken string            `json:"next_page_token,omitempty"`
}

type listCategoriesResponse struct {
	Categories []categoryResponse `json:"categories"`
}


type createProductRequest struct {
	Name        string   `json:"name"`
	Description string   `json:"description"`
	PricePoints int64    `json:"price_points"`
	CategoryID  string   `json:"category_id"`
	Sizes       []string `json:"sizes"`
	PhotoKeys   []string `json:"photo_keys"`
}

type updateProductRequest struct {
	Name        string   `json:"name"`
	Description string   `json:"description"`
	PricePoints int64    `json:"price_points"`
	CategoryID  string   `json:"category_id"`
	Sizes       []string `json:"sizes"`
	PhotoKeys   []string `json:"photo_keys"`
	Active      bool     `json:"active"`
	Version     int      `json:"version"`
}

type createCategoryRequest struct {
	Code string `json:"code"`
	Name string `json:"name"`
}

type updateCategoryRequest struct {
	Name   string `json:"name"`
	Active bool   `json:"active"`
}

func (s *Server) ListProducts(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	res, err := s.read.ListProducts(r.Context(), dto.ListProductsInput{
		CategoryID: q.Get("category_id"),
		Search:     q.Get("search"),
		ActiveOnly: parseBoolDefault(q.Get("active_only"), true),
		PageSize:   atoiOr(q.Get("page_size"), 0),
		PageToken:  q.Get("page_token"),
	})
	if err != nil {
		writeError(w, err)
		return
	}

	products := make([]productResponse, 0, len(res.Products))
	for _, p := range res.Products {
		products = append(products, toProductResponse(p))
	}
	writeJSON(w, http.StatusOK, listProductsResponse{Products: products, NextPageToken: res.NextPageToken})
}

func (s *Server) GetProduct(w http.ResponseWriter, r *http.Request) {
	view, err := s.read.GetProduct(r.Context(), r.PathValue("product_id"))
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, toProductResponse(view))
}

func (s *Server) ListCategories(w http.ResponseWriter, r *http.Request) {
	activeOnly := parseBoolDefault(r.URL.Query().Get("active_only"), true)
	views, err := s.read.ListCategories(r.Context(), activeOnly)
	if err != nil {
		writeError(w, err)
		return
	}
	categories := make([]categoryResponse, 0, len(views))
	for _, c := range views {
		categories = append(categories, toCategoryResponse(c))
	}
	writeJSON(w, http.StatusOK, listCategoriesResponse{Categories: categories})
}


func (s *Server) CreateProduct(w http.ResponseWriter, r *http.Request) {
	var req createProductRequest
	if !decodeJSON(w, r, &req) {
		return
	}
	view, err := s.write.CreateProduct(r.Context(), dto.CreateProductInput{
		Name:        req.Name,
		Description: req.Description,
		PricePoints: req.PricePoints,
		CategoryID:  req.CategoryID,
		Sizes:       req.Sizes,
		PhotoKeys:   req.PhotoKeys,
	})
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, toProductResponse(view))
}

func (s *Server) UpdateProduct(w http.ResponseWriter, r *http.Request) {
	var req updateProductRequest
	if !decodeJSON(w, r, &req) {
		return
	}
	view, err := s.write.UpdateProduct(r.Context(), dto.UpdateProductInput{
		ProductID:   r.PathValue("product_id"),
		Name:        req.Name,
		Description: req.Description,
		PricePoints: req.PricePoints,
		CategoryID:  req.CategoryID,
		Sizes:       req.Sizes,
		PhotoKeys:   req.PhotoKeys,
		Active:      req.Active,
		Version:     req.Version,
	})
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, toProductResponse(view))
}

func (s *Server) DeactivateProduct(w http.ResponseWriter, r *http.Request) {
	if err := s.write.DeactivateProduct(r.Context(), r.PathValue("product_id")); err != nil {
		writeError(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}


func (s *Server) CreateCategory(w http.ResponseWriter, r *http.Request) {
	var req createCategoryRequest
	if !decodeJSON(w, r, &req) {
		return
	}
	view, err := s.write.CreateCategory(r.Context(), dto.CreateCategoryInput{Code: req.Code, Name: req.Name})
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, toCategoryResponse(view))
}

func (s *Server) UpdateCategory(w http.ResponseWriter, r *http.Request) {
	var req updateCategoryRequest
	if !decodeJSON(w, r, &req) {
		return
	}
	view, err := s.write.UpdateCategory(r.Context(), dto.UpdateCategoryInput{
		ID:     r.PathValue("category_id"),
		Name:   req.Name,
		Active: req.Active,
	})
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, toCategoryResponse(view))
}


func toCategoryResponse(v dto.CategoryView) categoryResponse {
	return categoryResponse{
		ID:        v.ID,
		Code:      v.Code,
		Name:      v.Name,
		Active:    v.Active,
		CreatedAt: v.CreatedAt,
		UpdatedAt: v.UpdatedAt,
	}
}

func toProductResponse(v dto.ProductView) productResponse {
	sizes := v.Sizes
	if sizes == nil {
		sizes = []string{}
	}
	photoKeys := v.PhotoKeys
	if photoKeys == nil {
		photoKeys = []string{}
	}
	return productResponse{
		ID:          v.ID,
		Name:        v.Name,
		Description: v.Description,
		PricePoints: v.PricePoints,
		Category:    toCategoryResponse(v.Category),
		Sizes:       sizes,
		PhotoKeys:   photoKeys,
		Active:      v.Active,
		Version:     v.Version,
		CreatedAt:   v.CreatedAt,
		UpdatedAt:   v.UpdatedAt,
	}
}


func decodeJSON(w http.ResponseWriter, r *http.Request, dst any) bool {
	if err := json.NewDecoder(r.Body).Decode(dst); err != nil {
		writeJSON(w, http.StatusBadRequest, apiError{Code: "INVALID_JSON", Message: "invalid request body"})
		return false
	}
	return true
}

func parseBoolDefault(s string, fallback bool) bool {
	if s == "" {
		return fallback
	}
	v, err := strconv.ParseBool(s)
	if err != nil {
		return fallback
	}
	return v
}

func atoiOr(s string, fallback int) int {
	if s == "" {
		return fallback
	}
	n, err := strconv.Atoi(s)
	if err != nil {
		return fallback
	}
	return n
}
