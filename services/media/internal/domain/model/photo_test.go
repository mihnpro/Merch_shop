package model

import (
	"regexp"
	"testing"

	"github.com/stretchr/testify/assert"
)

var keyPattern = regexp.MustCompile(`^products/[0-9a-f-]{36}\.(jpg|png|webp)$`)

func TestNewPhotoKey(t *testing.T) {
	t.Run("format with extension", func(t *testing.T) {
		for _, ext := range []string{"jpg", "png", "webp"} {
			key := NewPhotoKey(ext)
			assert.Regexpf(t, keyPattern, key, "key %q", key)
		}
	})

	t.Run("keys are unique", func(t *testing.T) {
		seen := make(map[string]bool)
		for i := 0; i < 100; i++ {
			key := NewPhotoKey("jpg")
			assert.False(t, seen[key], "duplicate key %q", key)
			seen[key] = true
		}
	})
}
