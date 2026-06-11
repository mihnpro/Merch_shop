package kafka

import (
	"encoding/json"
	"testing"
)

func TestOrderCreatedEvent_Schema(t *testing.T) {
	const payload = `{
		"order_id": "11111111-1111-1111-1111-111111111111",
		"user_id":  "22222222-2222-2222-2222-222222222222",
		"items": [
			{"product_id": "33333333-3333-3333-3333-333333333333", "quantity": 2}
		]
	}`

	var ev OrderCreatedEvent
	if err := json.Unmarshal([]byte(payload), &ev); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if ev.OrderID != "11111111-1111-1111-1111-111111111111" {
		t.Errorf("order_id not mapped: %q", ev.OrderID)
	}
	if ev.UserID != "22222222-2222-2222-2222-222222222222" {
		t.Errorf("user_id not mapped: %q", ev.UserID)
	}
	if len(ev.Items) != 1 || ev.Items[0].ProductID != "33333333-3333-3333-3333-333333333333" || ev.Items[0].Quantity != 2 {
		t.Errorf("items not mapped: %+v", ev.Items)
	}
}
