package port

import "github.com/mihnpro/Merch_shop/services/media/internal/app/dto"

type Verifier interface {
	ParseAccess(token string) (dto.Identity, error)
}
