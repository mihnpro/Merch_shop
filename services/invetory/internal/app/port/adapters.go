package port

import "context"

type OrderNotifier interface {
	UpdateOrderStatus(ctx context.Context, orderID, newStatus, reason string) error
}
