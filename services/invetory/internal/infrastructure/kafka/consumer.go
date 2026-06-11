package kafka

import (
	"context"
	"encoding/json"

	"github.com/segmentio/kafka-go"
	"go.uber.org/zap"

	"github.com/mihnpro/Merch_shop/services/invetory/internal/app/dto"
	"github.com/mihnpro/Merch_shop/services/invetory/internal/app/service"
)


type OrderCreatedEvent struct {
	OrderID string      `json:"order_id"`
	UserID  string      `json:"user_id"`
	Items   []EventItem `json:"items"`
}

type EventItem struct {
	ProductID string `json:"product_id"`
	Quantity  int    `json:"quantity"`
}

type Consumer struct {
	reader *kafka.Reader
	svc    service.InventoryService
	logger *zap.Logger
}

func NewConsumer(brokers []string, svc service.InventoryService, logger *zap.Logger) *Consumer {
	r := kafka.NewReader(kafka.ReaderConfig{
		Brokers:  brokers,
		Topic:    "order.created",
		GroupID:  "inventory",
		MinBytes: 1,
		MaxBytes: 10e6,
	})
	return &Consumer{reader: r, svc: svc, logger: logger}
}

func (c *Consumer) Run(ctx context.Context) {
	defer c.reader.Close()
	for {
		msg, err := c.reader.FetchMessage(ctx)
		if err != nil {
			if ctx.Err() != nil {
				return
			}
			c.logger.Error("kafka: fetch message", zap.Error(err))
			continue
		}
		c.handle(ctx, msg)
	}
}

func (c *Consumer) handle(ctx context.Context, msg kafka.Message) {
	var ev OrderCreatedEvent
	if err := json.Unmarshal(msg.Value, &ev); err != nil {
		c.logger.Error("kafka: unmarshal event", zap.Error(err))
		_ = c.reader.CommitMessages(ctx, msg)
		return
	}

	items := make([]dto.ReserveItemInput, 0, len(ev.Items))
	for _, it := range ev.Items {
		items = append(items, dto.ReserveItemInput{
			ProductID: it.ProductID,
			Qty:       it.Quantity,
		})
	}

	if err := c.svc.HandleOrderCreated(ctx, dto.ReserveStockInput{OrderID: ev.OrderID, Items: items}); err != nil {
		c.logger.Error("kafka: handle order.created failed, will retry on redelivery",
			zap.String("order_id", ev.OrderID),
			zap.Error(err),
		)
		return
	}

	if err := c.reader.CommitMessages(ctx, msg); err != nil {
		c.logger.Error("kafka: commit offset failed",
			zap.String("order_id", ev.OrderID),
			zap.Error(err),
		)
	}
}
