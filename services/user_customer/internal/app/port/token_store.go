package port

import (
	"context"
	"time"
)

type TokenStore interface {
	Store(ctx context.Context, userID string, token string, ttl time.Duration) error
	GetUserID(ctx context.Context, token string) (string, error)
	Delete(ctx context.Context, token string) error
	DeleteByUserID(ctx context.Context, userID string) error
}
