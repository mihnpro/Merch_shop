package dto

type ProductInfo struct {
	Name        string
	PricePoints int
	Active      bool
}

type CartItemView struct {
	ID           string `json:"id"`
	ProductID    string `json:"product_id"`
	ProductName  string `json:"product_name"`
	Quantity     int    `json:"quantity"`
	PricePerUnit int    `json:"price_per_unit"`
	LineTotal    int    `json:"line_total"`
}

type CartView struct {
	Items     []CartItemView `json:"items"`
	Total     int            `json:"total"`
	ItemCount int            `json:"item_count"`
}

type GetCartInput struct {
	UserID string
}

type AddItemInput struct {
	UserID    string
	ProductID string
	Quantity  int
}

type UpdateItemInput struct {
	UserID      string
	ItemID      string
	NewQuantity int
}

type RemoveItemInput struct {
	UserID string
	ItemID string
}

type ClearCartInput struct {
	UserID string
}
