package model

import (
	"fmt"

	"github.com/google/uuid"
)

const MaxPhotoSize = 5 * 1024 * 1024

const keyPrefix = "products"

func NewPhotoKey(ext string) string {
	return fmt.Sprintf("%s/%s.%s", keyPrefix, uuid.NewString(), ext)
}
