package events

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// canonicalPayload is the exact wire format of an order.created event. The
// inventory consumer is tested against the same shape; together these pin the
// cross-service contract.
const canonicalPayload = `{"order_id":"11111111-1111-1111-1111-111111111111","user_id":"22222222-2222-2222-2222-222222222222","items":[{"product_id":"33333333-3333-3333-3333-333333333333","quantity":2}]}`

func TestTypeConstant(t *testing.T) {
	assert.Equal(t, "order.created", TypeOrderCreated)
}

func TestOrderCreatedMarshalContract(t *testing.T) {
	ev := OrderCreated{
		OrderID: "11111111-1111-1111-1111-111111111111",
		UserID:  "22222222-2222-2222-2222-222222222222",
		Items: []OrderCreatedItem{
			{ProductID: "33333333-3333-3333-3333-333333333333", Quantity: 2},
		},
	}
	out, err := json.Marshal(ev)
	require.NoError(t, err)
	assert.JSONEq(t, canonicalPayload, string(out))
}

func TestOrderCreatedRoundTrip(t *testing.T) {
	ev := OrderCreated{
		OrderID: "o-1",
		UserID:  "u-1",
		Items: []OrderCreatedItem{
			{ProductID: "p-1", Quantity: 3},
			{ProductID: "p-2", Quantity: 1},
		},
	}
	raw, err := json.Marshal(ev)
	require.NoError(t, err)

	var decoded OrderCreated
	require.NoError(t, json.Unmarshal(raw, &decoded))
	assert.Equal(t, ev, decoded)
	require.Len(t, decoded.Items, 2)
	assert.Equal(t, "p-2", decoded.Items[1].ProductID)
	assert.Equal(t, 1, decoded.Items[1].Quantity)
}
