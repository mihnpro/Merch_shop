package service

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"

	"github.com/mihnpro/Merch_shop/services/media/internal/app/dto"
	"github.com/mihnpro/Merch_shop/services/media/internal/app/port"
	"github.com/mihnpro/Merch_shop/services/media/internal/domain"
	"github.com/mihnpro/Merch_shop/services/media/internal/domain/model"
	vo "github.com/mihnpro/Merch_shop/services/media/internal/domain/valueobject"
)

type Uploader interface {
	Upload(ctx context.Context, in dto.UploadInput) (dto.UploadResult, error)
}

type uploader struct {
	storage port.Storage
}

func NewUploader(storage port.Storage) Uploader {
	return &uploader{storage: storage}
}

func (u *uploader) Upload(ctx context.Context, in dto.UploadInput) (dto.UploadResult, error) {
	ct, err := vo.NewContentType(in.ContentType)
	if err != nil {
		return dto.UploadResult{}, err
	}
	if in.Body == nil {
		return dto.UploadResult{}, domain.ErrEmptyRequest
	}
	if in.Size > model.MaxPhotoSize {
		return dto.UploadResult{}, domain.ErrFileTooLarge
	}

	guarded := &limitReader{r: in.Body, remaining: model.MaxPhotoSize}

	head := make([]byte, 1)
	n, err := io.ReadFull(guarded, head)
	switch {
	case errors.Is(err, io.EOF):
		return dto.UploadResult{}, domain.ErrEmptyFile
	case errors.Is(err, domain.ErrFileTooLarge):
		return dto.UploadResult{}, domain.ErrFileTooLarge
	case err != nil && !errors.Is(err, io.ErrUnexpectedEOF):
		return dto.UploadResult{}, fmt.Errorf("%w: read body: %v", domain.ErrStorageUnavailable, err)
	}

	body := io.MultiReader(bytes.NewReader(head[:n]), guarded)

	key := model.NewPhotoKey(ct.Ext())
	size := in.Size
	if size <= 0 {
		size = -1
	}

	if err := u.storage.Put(ctx, key, ct.String(), body, size); err != nil {
		if errors.Is(err, domain.ErrFileTooLarge) {
			return dto.UploadResult{}, domain.ErrFileTooLarge
		}
		return dto.UploadResult{}, fmt.Errorf("%w: put object: %v", domain.ErrStorageUnavailable, err)
	}
	return dto.UploadResult{PhotoKey: key}, nil
}

type limitReader struct {
	r         io.Reader
	remaining int64
}

func (l *limitReader) Read(p []byte) (int, error) {
	if l.remaining < 0 {
		return 0, domain.ErrFileTooLarge
	}

	if int64(len(p)) > l.remaining+1 {
		p = p[:l.remaining+1]
	}
	n, err := l.r.Read(p)
	l.remaining -= int64(n)
	if l.remaining < 0 {
		return n, domain.ErrFileTooLarge
	}
	return n, err
}
