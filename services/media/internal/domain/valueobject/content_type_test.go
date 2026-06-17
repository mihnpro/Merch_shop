package valueobject

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/mihnpro/Merch_shop/services/media/internal/domain"
)

func TestNewContentType(t *testing.T) {
	cases := map[string]struct {
		raw string
		ext string
	}{
		"jpeg":              {"image/jpeg", "jpg"},
		"png":               {"image/png", "png"},
		"webp":              {"image/webp", "webp"},
		"uppercase":         {"IMAGE/JPEG", "jpg"},
		"with charset":      {"image/png; charset=utf-8", "png"},
		"surrounding space": {"  image/webp  ", "webp"},
	}

	for name, c := range cases {
		t.Run("valid "+name, func(t *testing.T) {
			ct, err := NewContentType(c.raw)
			require.NoError(t, err)
			assert.Equal(t, c.ext, ct.Ext())
		})
	}

	for name, raw := range map[string]string{
		"gif":         "image/gif",
		"pdf":         "application/pdf",
		"empty":       "",
		"plain image": "image",
	} {
		t.Run("invalid "+name, func(t *testing.T) {
			_, err := NewContentType(raw)
			assert.ErrorIs(t, err, domain.ErrInvalidContentType)
		})
	}

	t.Run("mime preserved in String", func(t *testing.T) {
		ct, err := NewContentType("image/jpeg")
		require.NoError(t, err)
		assert.Equal(t, "image/jpeg", ct.String())
	})
}
