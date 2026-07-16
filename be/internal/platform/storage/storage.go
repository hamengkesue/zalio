package storage

import (
	"context"
	"fmt"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"

	"zalio-erp-be/internal/platform/config"
)

// Storage membungkus client MinIO + nama bucket.
type Storage struct {
	Client *minio.Client
	Bucket string
}

// New membuat client MinIO dan memastikan bucket-nya ada.
func New(cfg *config.Config) (*Storage, error) {
	client, err := minio.New(cfg.MinioEndpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(cfg.MinioAccessKey, cfg.MinioSecretKey, ""),
		Secure: cfg.MinioUseSSL,
	})
	if err != nil {
		return nil, fmt.Errorf("minio init: %w", err)
	}

	ctx := context.Background()
	exists, err := client.BucketExists(ctx, cfg.MinioBucket)
	if err != nil {
		return nil, fmt.Errorf("minio bucket check: %w", err)
	}
	if !exists {
		if err := client.MakeBucket(ctx, cfg.MinioBucket, minio.MakeBucketOptions{}); err != nil {
			return nil, fmt.Errorf("minio make bucket: %w", err)
		}
	}

	return &Storage{Client: client, Bucket: cfg.MinioBucket}, nil
}
