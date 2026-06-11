package repository

import (
	"context"
	"time"

	"github.com/google/uuid"

	"github.com/mihnpro/Merch_shop/services/order/internal/app/outbox"
)

type OutboxRepository interface {
	FetchUnsent(ctx context.Context, limit int) ([]outbox.Event, error)
	MarkSent(ctx context.Context, id uuid.UUID, sentAt time.Time) error
}
