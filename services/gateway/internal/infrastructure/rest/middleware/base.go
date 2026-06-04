package middleware

import (
	"go.uber.org/zap"

	"github.com/mihnpro/Merch_shop/services/gateway/internal/app/port"
)

type Middleware struct {
	log            *zap.Logger
	verify         port.Verifier
	publicPrefixes []string
	corsOrigins    []string
}

func NewMiddleware(log *zap.Logger, verify port.Verifier, publicPrefixes, corsOrigins []string) *Middleware {
	return &Middleware{
		log:            log,
		verify:         verify,
		publicPrefixes: publicPrefixes,
		corsOrigins:    corsOrigins,
	}
}
