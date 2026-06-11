package psql

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"

	"github.com/mihnpro/Merch_shop/services/order/internal/app/outbox"
)

type OutboxRepository struct {
	db *sqlx.DB
}

func NewOutboxRepository(db *sqlx.DB) *OutboxRepository {
	return &OutboxRepository{db: db}
}

type outboxRow struct {
	ID        uuid.UUID `db:"id"`
	EventType string    `db:"event_type"`
	Payload   []byte    `db:"payload"`
	CreatedAt time.Time `db:"created_at"`
}

func (r *OutboxRepository) Add(ctx context.Context, exec sqlx.ExecerContext, eventType string, payload []byte) error {
	_, err := exec.ExecContext(ctx,
		`INSERT INTO outbox (id, event_type, payload, created_at) VALUES (gen_random_uuid(), $1, $2, NOW())`,
		eventType, payload,
	)
	if err != nil {
		return fmt.Errorf("insert outbox: %w", err)
	}
	return nil
}

func (r *OutboxRepository) FetchUnsent(ctx context.Context, limit int) ([]outbox.Event, error) {
	var rows []outboxRow
	err := r.db.SelectContext(ctx, &rows,
		`SELECT id, event_type, payload, created_at FROM outbox WHERE sent_at IS NULL ORDER BY created_at LIMIT $1`,
		limit)
	if err != nil {
		return nil, err
	}
	events := make([]outbox.Event, 0, len(rows))
	for _, row := range rows {
		events = append(events, outbox.Event{
			ID:        row.ID,
			EventType: row.EventType,
			Payload:   row.Payload,
			CreatedAt: row.CreatedAt,
		})
	}
	return events, nil
}

func (r *OutboxRepository) MarkSent(ctx context.Context, id uuid.UUID, sentAt time.Time) error {
	_, err := r.db.ExecContext(ctx, `UPDATE outbox SET sent_at = $1 WHERE id = $2`, sentAt, id)
	return err
}
