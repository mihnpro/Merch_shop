package valueobject

import (
	"strings"

	"github.com/mihnpro/Merch_shop/services/user_customer/internal/domain"
)

type Status string

const (
	StatusActive  Status = "active"
	StatusBlocked Status = "blocked"
)

func NewStatus(raw string) (Status, error) {
	switch s := Status(strings.TrimSpace(raw)); s {
	case StatusActive, StatusBlocked:
		return s, nil
	default:
		return "", domain.ErrInvalidStatus
	}
}

func NewStatusFromStored(v string) Status { return Status(v) }

func (s Status) String() string { return string(s) }

func (s Status) IsBlocked() bool { return s == StatusBlocked }
