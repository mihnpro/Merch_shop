package port

import "context"

type CartItem struct {
	ItemID      string
	ProductID   string
	ProductName string
	Quantity    int
	PricePoints int64
}


type CartClient interface {
	GetCart(ctx context.Context, userID string) ([]CartItem, error)
	ClearCart(ctx context.Context, userID string) error
}


type UserClient interface {
	GetBalance(ctx context.Context, userID string) (int64, error)
	DeductPoints(ctx context.Context, userID string, amount int64, operationID, orderID, reason string) error
	AddPoints(ctx context.Context, userID string, amount int64, operationID, orderID, reason string) error
}


type InventoryClient interface {
	ReleaseReserve(ctx context.Context, orderID, reason string) error
}
