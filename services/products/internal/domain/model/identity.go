package model

import "github.com/google/uuid"


type Identity struct {
	UserID uuid.UUID
	Email  string
	Role   string
}
