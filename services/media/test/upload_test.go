package e2e

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestUploadAuthorization(t *testing.T) {
	t.Run("no token → 401", func(t *testing.T) {
		status, raw, err := newClient(cfg).uploadPhoto("file", "image/png", samplePNG)
		require.NoError(t, err)
		requireStatus(t, http.StatusUnauthorized, status, raw)
	})

	t.Run("non-admin → 403", func(t *testing.T) {
		user := newUserClient(t)
		status, raw, err := user.uploadPhoto("file", "image/png", samplePNG)
		require.NoError(t, err)
		requireStatus(t, http.StatusForbidden, status, raw)
	})
}

func TestUploadPhoto(t *testing.T) {
	admin := adminClient(t)

	t.Run("happy path returns photo key", func(t *testing.T) {
		status, raw, err := admin.uploadPhoto("file", "image/png", samplePNG)
		require.NoError(t, err)
		requireStatus(t, http.StatusOK, status, raw)
		key := decode[uploadView](t, raw).PhotoKey
		assert.Regexp(t, `^products/[0-9a-f-]{36}\.png$`, key)
	})

	t.Run("unsupported content type → 400", func(t *testing.T) {
		status, raw, err := admin.uploadPhoto("file", "application/pdf", []byte("%PDF-1.4"))
		require.NoError(t, err)
		requireStatus(t, http.StatusBadRequest, status, raw)
	})

	t.Run("empty file → 400", func(t *testing.T) {
		status, raw, err := admin.uploadPhoto("file", "image/png", nil)
		require.NoError(t, err)
		requireStatus(t, http.StatusBadRequest, status, raw)
	})

	t.Run("missing file field → 400", func(t *testing.T) {
		status, raw, err := admin.uploadPhoto("not_file", "image/png", samplePNG)
		require.NoError(t, err)
		requireStatus(t, http.StatusBadRequest, status, raw)
	})
}
