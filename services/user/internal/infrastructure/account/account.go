package account

import (
	"time"

	"github.com/mihnpro/Merch_shop/services/user_customer/internal/app/port"
)

type account struct {
	accessSecret  string
	refreshSecret string
	accessTTL     time.Duration
	refreshTTL    time.Duration
}

func NewAccount(accessSecret, refreshSecret string, accessTTL, refreshTTL time.Duration) port.Account {
	return &account{
		accessSecret:  accessSecret,
		refreshSecret: refreshSecret,
		accessTTL:     accessTTL,
		refreshTTL:    refreshTTL,
	}
}
