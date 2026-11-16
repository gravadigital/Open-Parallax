package storage

import (
	"context"
	"fmt"
	"time"

	"github.com/gravadigital/telescopio-api/internal/config"
)

// NewFileStorage creates a new FileStorage instance based on configuration
func NewFileStorage(cfg *config.Config) (FileStorage, error) {
	switch cfg.Storage.Provider {
	case "local":
		return NewLocalStorage(cfg.Storage.LocalPath)

	case "s3":
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()
		return NewS3Storage(ctx, S3Config{
			Endpoint:        cfg.Storage.S3Endpoint,
			PresignEndpoint: cfg.Storage.S3PresignEndpoint,
			Region:          cfg.Storage.S3Region,
			Bucket:          cfg.Storage.S3Bucket,
			AccessKeyID:     cfg.Storage.S3AccessKeyID,
			SecretAccessKey: cfg.Storage.S3SecretAccessKey,
		})

	case "minio":
		return nil, fmt.Errorf("STORAGE_PROVIDER=minio was removed: use STORAGE_PROVIDER=s3 with the S3_* variables (see documentation/configuration.md)")

	default:
		return nil, fmt.Errorf("unsupported storage provider: %s (must be 'local' or 's3')", cfg.Storage.Provider)
	}
}
