package valueobject

import "github.com/mihnpro/Merch_shop/services/products/internal/domain"

type PricePoints struct {
	value int64
}

func NewPricePoints(v int64) (PricePoints, error) {
	if v <= 0 {
		return PricePoints{}, domain.ErrInvalidPrice
	}
	return PricePoints{value: v}, nil
}

func NewPricePointsFromStored(v int64) PricePoints {
	return PricePoints{value: v}
}

func (p PricePoints) Int64() int64 { return p.value }
