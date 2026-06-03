package middleware

import (
	"github.com/mihnpro/Merch_shop/services/gateway/internal/app/port"
	"go.uber.org/zap"
)

type Middleware struct {
	log         *zap.Logger
	verify      port.Verifier
	corsOrigins []string
}

func NewMiddleware(log *zap.Logger, verify port.Verifier, corsOrigins []string) *Middleware {
	return &Middleware{log: log, verify: verify, corsOrigins: corsOrigins}
}
