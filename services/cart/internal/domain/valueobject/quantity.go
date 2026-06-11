package valueobject

import (
	"encoding/json"

	"github.com/mihnpro/Merch_shop/services/cart/internal/domain"
)

const (
	minQty = 1
	maxQty = 9999
)

type Quantity struct {
	value int
}

func NewQuantity(v int) (Quantity, error) {
	if v < minQty || v > maxQty {
		return Quantity{}, domain.ErrInvalidQuantity
	}
	return Quantity{value: v}, nil
}

func NewQuantityFromStored(v int) Quantity {
	return Quantity{value: v}
}

func (q Quantity) Int() int { return q.value }

func (q Quantity) MarshalJSON() ([]byte, error) { return json.Marshal(q.value) }

func (q *Quantity) UnmarshalJSON(data []byte) error {
	var v int
	if err := json.Unmarshal(data, &v); err != nil {
		return err
	}
	q.value = v
	return nil
}
