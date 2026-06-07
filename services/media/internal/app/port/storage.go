package port

import (
	"context"
	"io"
)

type Storage interface {
	Put(ctx context.Context, key, contentType string, r io.Reader, size int64) error
}
