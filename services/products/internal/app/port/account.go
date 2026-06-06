package port

import "github.com/mihnpro/Merch_shop/services/products/internal/domain/model"



type TokenValidator interface {
	ValidateToken(token string) (model.Identity, error)
}
