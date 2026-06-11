package events

const TypeOrderCreated = "order.created"

type OrderCreated struct {
	OrderID string             `json:"order_id"`
	UserID  string             `json:"user_id"`
	Items   []OrderCreatedItem `json:"items"`
}

type OrderCreatedItem struct {
	ProductID string `json:"product_id"`
	Quantity  int    `json:"quantity"`
}
