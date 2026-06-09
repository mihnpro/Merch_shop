package valueobject

import "github.com/mihnpro/Merch_shop/services/invetory/internal/domain"


type Delta struct {
	value int
}

func NewDelta(v int) (Delta, error) {
	if v == 0 {
		return Delta{}, domain.ErrZeroDelta
	}
	return Delta{value: v}, nil
}

func NewDeltaFromStored(v int) Delta {
	return Delta{value: v}
}

func (d Delta) Int() int { return d.value }
