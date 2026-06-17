package service

import (
	"context"
	"errors"
	"io"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/mihnpro/Merch_shop/services/media/internal/app/dto"
	"github.com/mihnpro/Merch_shop/services/media/internal/domain"
	"github.com/mihnpro/Merch_shop/services/media/internal/domain/model"
)

var errBoom = errors.New("boom")

type fakeStorage struct {
	putErr      error
	key         string
	contentType string
	size        int64
	written     int64
	called      bool
}

func (f *fakeStorage) Put(_ context.Context, key, contentType string, r io.Reader, size int64) error {
	f.called = true
	f.key, f.contentType, f.size = key, contentType, size
	n, err := io.Copy(io.Discard, r)
	f.written = n
	if err != nil {
		return err
	}
	return f.putErr
}

func TestUpload(t *testing.T) {
	ctx := context.Background()

	t.Run("success returns generated key", func(t *testing.T) {
		store := &fakeStorage{}
		u := NewUploader(store)
		res, err := u.Upload(ctx, dto.UploadInput{
			ContentType: "image/jpeg", Size: 5, Body: strings.NewReader("hello"),
		})
		require.NoError(t, err)
		assert.Regexp(t, `^products/[0-9a-f-]{36}\.jpg$`, res.PhotoKey)
		assert.Equal(t, res.PhotoKey, store.key)
		assert.Equal(t, "image/jpeg", store.contentType)
		assert.Equal(t, int64(5), store.size)
		assert.EqualValues(t, 5, store.written)
	})

	t.Run("unknown size passed as -1", func(t *testing.T) {
		store := &fakeStorage{}
		u := NewUploader(store)
		_, err := u.Upload(ctx, dto.UploadInput{ContentType: "image/png", Size: -1, Body: strings.NewReader("data")})
		require.NoError(t, err)
		assert.Equal(t, int64(-1), store.size)
	})

	t.Run("invalid content type", func(t *testing.T) {
		u := NewUploader(&fakeStorage{})
		_, err := u.Upload(ctx, dto.UploadInput{ContentType: "application/pdf", Size: 1, Body: strings.NewReader("x")})
		assert.ErrorIs(t, err, domain.ErrInvalidContentType)
	})

	t.Run("nil body", func(t *testing.T) {
		u := NewUploader(&fakeStorage{})
		_, err := u.Upload(ctx, dto.UploadInput{ContentType: "image/jpeg", Size: 1, Body: nil})
		assert.ErrorIs(t, err, domain.ErrEmptyRequest)
	})

	t.Run("size pre-check too large", func(t *testing.T) {
		u := NewUploader(&fakeStorage{})
		_, err := u.Upload(ctx, dto.UploadInput{ContentType: "image/jpeg", Size: model.MaxPhotoSize + 1, Body: strings.NewReader("x")})
		assert.ErrorIs(t, err, domain.ErrFileTooLarge)
	})

	t.Run("empty file", func(t *testing.T) {
		u := NewUploader(&fakeStorage{})
		_, err := u.Upload(ctx, dto.UploadInput{ContentType: "image/jpeg", Size: 0, Body: strings.NewReader("")})
		assert.ErrorIs(t, err, domain.ErrEmptyFile)
	})

	t.Run("oversized stream rejected during put", func(t *testing.T) {
		store := &fakeStorage{}
		u := NewUploader(store)
		big := strings.NewReader(strings.Repeat("a", model.MaxPhotoSize+10))
		_, err := u.Upload(ctx, dto.UploadInput{ContentType: "image/jpeg", Size: -1, Body: big})
		assert.ErrorIs(t, err, domain.ErrFileTooLarge)
	})

	t.Run("storage file-too-large mapped", func(t *testing.T) {
		store := &fakeStorage{putErr: domain.ErrFileTooLarge}
		u := NewUploader(store)
		_, err := u.Upload(ctx, dto.UploadInput{ContentType: "image/jpeg", Size: 3, Body: strings.NewReader("abc")})
		assert.ErrorIs(t, err, domain.ErrFileTooLarge)
	})

	t.Run("storage error wrapped as unavailable", func(t *testing.T) {
		store := &fakeStorage{putErr: errBoom}
		u := NewUploader(store)
		_, err := u.Upload(ctx, dto.UploadInput{ContentType: "image/jpeg", Size: 3, Body: strings.NewReader("abc")})
		assert.ErrorIs(t, err, domain.ErrStorageUnavailable)
	})
}
