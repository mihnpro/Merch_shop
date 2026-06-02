package port

import "github.com/mihnpro/Merch_shop/services/gateway/internal/domain"

type Verifier interface {
	VerifyAccess(token string) (*domain.Claims, error)
	VerifyRefresh(token string) (*domain.Claims, error)
}
