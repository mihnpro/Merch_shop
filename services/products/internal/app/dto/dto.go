package dto

import "time"

type CreateProductInput struct {
	Name        string
	Description string
	PricePoints int64
	CategoryID  string
	Sizes       []string
	PhotoKey    string
}

type UpdateProductInput struct {
	ProductID   string
	Name        string
	Description string
	PricePoints int64
	CategoryID  string
	Sizes       []string
	PhotoKey    string
	Active      bool
	Version     int
}

type ListProductsInput struct {
	CategoryID string
	Search     string
	ActiveOnly bool
	PageSize   int
	PageToken  string
}

type CreateCategoryInput struct {
	Code string
	Name string
}

type UpdateCategoryInput struct {
	ID     string
	Name   string
	Active bool
}


type CategoryView struct {
	ID        string
	Code      string
	Name      string
	Active    bool
	CreatedAt time.Time
	UpdatedAt time.Time
}

type ProductView struct {
	ID          string
	Name        string
	Description string
	PricePoints int64
	Category    CategoryView
	Sizes       []string
	PhotoKey    string
	Active      bool
	Version     int
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

type ListProductsResult struct {
	Products      []ProductView
	NextPageToken string
}
