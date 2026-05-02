package storage

import "context"

// CVStorage defines the contract for CV file operations.
// This interface is what services depend on — not the concrete S3 struct.
type CVStorage interface {
	GenerateUploadURL(ctx context.Context, key string) (string, error)
	GenerateDownloadURL(ctx context.Context, key string) (string, error)
	DeleteObject(ctx context.Context, key string) error
}
