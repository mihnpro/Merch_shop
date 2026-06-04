package model

import "github.com/google/uuid"

type Role struct {
	ID          uuid.UUID
	Code        string
	Name        string
	Description string
}
