package kafka

import (
	"context"
	"time"

	"github.com/segmentio/kafka-go"
	"go.uber.org/zap"

	"github.com/mihnpro/Merch_shop/services/order/internal/domain/repository"
)

type OutboxWorker struct {
	repo   repository.OutboxRepository
	writer *kafka.Writer
	logger *zap.Logger
}

func NewOutboxWorker(repo repository.OutboxRepository, brokers []string, topic string, logger *zap.Logger) *OutboxWorker {
	w := &kafka.Writer{
		Addr:         kafka.TCP(brokers...),
		Topic:        topic,
		Balancer:     &kafka.LeastBytes{},
		RequiredAcks: kafka.RequireAll,
	}
	return &OutboxWorker{repo: repo, writer: w, logger: logger}
}

func (w *OutboxWorker) Run(ctx context.Context) {
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			_ = w.writer.Close()
			return
		case <-ticker.C:
			w.process(ctx)
		}
	}
}

func (w *OutboxWorker) process(ctx context.Context) {
	events, err := w.repo.FetchUnsent(ctx, 50)
	if err != nil {
		w.logger.Error("outbox: fetch unsent", zap.Error(err))
		return
	}
	for _, ev := range events {
		msg := kafka.Message{
			Key:   []byte(ev.ID.String()),
			Value: ev.Payload,
			Headers: []kafka.Header{
				{Key: "event_type", Value: []byte(ev.EventType)},
			},
		}
		if err := w.writer.WriteMessages(ctx, msg); err != nil {
			w.logger.Error("outbox: publish failed", zap.String("event_id", ev.ID.String()), zap.Error(err))
			continue
		}
		if err := w.repo.MarkSent(ctx, ev.ID, time.Now()); err != nil {
			w.logger.Error("outbox: mark sent failed", zap.String("event_id", ev.ID.String()), zap.Error(err))
		}
	}
}
