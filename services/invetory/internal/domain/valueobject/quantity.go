package valueobject

import "github.com/mihnpro/Merch_shop/services/invetory/internal/domain"

const (
	minQty = 1
	maxQty = 100
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
