package dto

import "time"

type OrderItemView struct {
	ID          string `json:"id"`
	ProductID   string `json:"product_id"`
	ProductName string `json:"product_name"`
	Quantity    int    `json:"quantity"`
	PricePoints int64  `json:"price_points"`
}

type OrderView struct {
	ID              string          `json:"id"`
	UserID          string          `json:"user_id"`
	Status          string          `json:"status"`
	TotalPoints     int64           `json:"total_points"`
	Note            *string         `json:"note,omitempty"`
	CancelReason    *string         `json:"cancel_reason,omitempty"`
	DeliveryAddress string          `json:"delivery_address"`
	Items           []OrderItemView `json:"items"`
	CreatedAt       time.Time       `json:"created_at"`
	UpdatedAt       time.Time       `json:"updated_at"`
}

type CreateOrderInput struct {
	UserID          string
	DeliveryAddress string
	Note            *string
}

type UpdateOrderStatusInput struct {
	OrderID   string
	NewStatus string
	Reason    string
}

type GetOrderInput struct {
	OrderID string
}

type ListOrdersInput struct {
	UserID    string
	Status    string
	PageSize  int
	PageToken string
}

type TopProductView struct {
	ProductID   string `json:"product_id"`
	ProductName string `json:"product_name"`
	Quantity    int    `json:"quantity"`
}

type AnalyticsView struct {
	Period            string           `json:"period"`
	OrdersCount       int              `json:"orders_count"`
	PointsSpent       int64            `json:"points_spent"`
	AverageOrderValue int64            `json:"average_order_value"`
	TopProducts       []TopProductView `json:"top_products"`
}
