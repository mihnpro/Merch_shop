package outbox

import (
	"time"

	"github.com/google/uuid"
)

type Event struct {
	ID        uuid.UUID
	EventType string
	Payload   []byte
	CreatedAt time.Time
	SentAt    *time.Time
}
