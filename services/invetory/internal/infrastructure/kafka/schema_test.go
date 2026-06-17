package kafka

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)
const canonicalPayload = `{"order_id":"11111111-1111-1111-1111-111111111111","user_id":"22222222-2222-2222-2222-222222222222","items":[{"product_id":"33333333-3333-3333-3333-333333333333","quantity":2}]}`

func TestOrderCreatedEvent_Contract(t *testing.T) {
	var ev OrderCreatedEvent
	require.NoError(t, json.Unmarshal([]byte(canonicalPayload), &ev))
	assert.Equal(t, "11111111-1111-1111-1111-111111111111", ev.OrderID)
	assert.Equal(t, "22222222-2222-2222-2222-222222222222", ev.UserID)
	require.Len(t, ev.Items, 1)
	assert.Equal(t, "33333333-3333-3333-3333-333333333333", ev.Items[0].ProductID)
	assert.Equal(t, 2, ev.Items[0].Quantity)
}

func TestOrderCreatedEvent_MultipleItems(t *testing.T) {
	const payload = `{"order_id":"o","user_id":"u","items":[{"product_id":"p1","quantity":2},{"product_id":"p2","quantity":5}]}`
	var ev OrderCreatedEvent
	require.NoError(t, json.Unmarshal([]byte(payload), &ev))
	require.Len(t, ev.Items, 2)
	assert.Equal(t, "p2", ev.Items[1].ProductID)
	assert.Equal(t, 5, ev.Items[1].Quantity)
}

func TestOrderCreatedEvent_EmptyItems(t *testing.T) {
	const payload = `{"order_id":"o","user_id":"u","items":[]}`
	var ev OrderCreatedEvent
	require.NoError(t, json.Unmarshal([]byte(payload), &ev))
	assert.Empty(t, ev.Items)
}

func TestOrderCreatedEvent_ForwardCompatible(t *testing.T) {
	const payload = `{"order_id":"o","user_id":"u","note":"new field","items":[{"product_id":"p","quantity":1,"size":"L"}]}`
	var ev OrderCreatedEvent
	require.NoError(t, json.Unmarshal([]byte(payload), &ev))
	assert.Equal(t, "o", ev.OrderID)
	require.Len(t, ev.Items, 1)
	assert.Equal(t, 1, ev.Items[0].Quantity)
}

func TestOrderCreatedEvent_InvalidJSON(t *testing.T) {
	var ev OrderCreatedEvent
	assert.Error(t, json.Unmarshal([]byte(`{"order_id":`), &ev))
}

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
