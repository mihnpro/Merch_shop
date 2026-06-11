package valueobject

import (
	"strings"

	"github.com/mihnpro/Merch_shop/services/user_customer/internal/domain"
)

type Role string

const (
	RoleUser  Role = "user"
	RoleAdmin Role = "admin"
)

func NewRole(raw string) (Role, error) {
	switch r := Role(strings.TrimSpace(raw)); r {
	case RoleUser, RoleAdmin:
		return r, nil
	default:
		return "", domain.ErrInvalidRole
	}
}

func NewRoleFromStored(v string) Role { return Role(v) }

func (r Role) String() string { return string(r) }

func (r Role) IsAdmin() bool { return r == RoleAdmin }
