package port

import "github.com/mihnpro/Merch_shop/services/invetory/internal/domain/model"


type TokenValidator interface {
	ValidateToken(token string) (model.Identity, error)
}
