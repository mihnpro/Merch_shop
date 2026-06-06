package minio

import (
	"context"
	"fmt"
	"io"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"

	"github.com/mihnpro/Merch_shop/services/media/internal/app/port"
)

type Config struct {
	Endpoint  string
	AccessKey string
	SecretKey string
	UseSSL    bool
	Bucket    string
}

type Storage struct {
	client *minio.Client
	bucket string
}

func New(cfg Config) (*Storage, error) {
	client, err := minio.New(cfg.Endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(cfg.AccessKey, cfg.SecretKey, ""),
		Secure: cfg.UseSSL,
	})
	if err != nil {
		return nil, fmt.Errorf("init minio client: %w", err)
	}
	return &Storage{client: client, bucket: cfg.Bucket}, nil
}

func (s *Storage) EnsureBucket(ctx context.Context) error {
	exists, err := s.client.BucketExists(ctx, s.bucket)
	if err != nil {
		return fmt.Errorf("check bucket %q: %w", s.bucket, err)
	}
	if !exists {
		if err := s.client.MakeBucket(ctx, s.bucket, minio.MakeBucketOptions{}); err != nil {
			return fmt.Errorf("create bucket %q: %w", s.bucket, err)
		}
	}
	if err := s.client.SetBucketPolicy(ctx, s.bucket, publicReadPolicy(s.bucket)); err != nil {
		return fmt.Errorf("set public-read policy on %q: %w", s.bucket, err)
	}
	return nil
}

func (s *Storage) Put(ctx context.Context, key, contentType string, r io.Reader, size int64) error {
	_, err := s.client.PutObject(ctx, s.bucket, key, r, size, minio.PutObjectOptions{
		ContentType: contentType,
	})
	if err != nil {
		return fmt.Errorf("put object %q: %w", key, err)
	}
	return nil
}

func publicReadPolicy(bucket string) string {
	return fmt.Sprintf(`{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Effect": "Allow",
      "Principal": {"AWS": ["*"]},
      "Action": ["s3:GetObject"],
      "Resource": ["arn:aws:s3:::%s/*"]
    }
  ]
}`, bucket)
}

var _ port.Storage = (*Storage)(nil)
