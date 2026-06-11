package port

import "github.com/mihnpro/Merch_shop/services/cart/internal/domain/model"

type TokenValidator interface {
	ValidateToken(token string) (model.Identity, error)
}
