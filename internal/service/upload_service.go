package service

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/gpslakshan/hireflow/internal/domain/dto"
	"github.com/gpslakshan/hireflow/internal/storage"
)

var ErrInvalidFileType = errors.New("only PDF files are accepted")

type UploadService struct {
	cvStorage storage.CVStorage
}

func NewUploadService(cvStorage storage.CVStorage) *UploadService {
	return &UploadService{cvStorage: cvStorage}
}

// GenerateCVUploadURL generates a pre-signed S3 URL for uploading a CV.
// The S3 key is scoped to the candidate: cvs/{candidateID}/{filename}
func (s *UploadService) GenerateCVUploadURL(ctx context.Context, candidateID string, fileName string) (dto.CVUploadURLResponse, error) {
	// Sanitise filename — strip path components and spaces
	sanitised := sanitiseFileName(fileName)
	if sanitised == "" {
		return dto.CVUploadURLResponse{}, fmt.Errorf("invalid file name")
	}

	// Enforce PDF only
	if !strings.HasSuffix(strings.ToLower(sanitised), ".pdf") {
		return dto.CVUploadURLResponse{}, ErrInvalidFileType
	}

	key := fmt.Sprintf("cvs/%s/%s", candidateID, sanitised)

	url, err := s.cvStorage.GenerateUploadURL(ctx, key)
	if err != nil {
		return dto.CVUploadURLResponse{}, fmt.Errorf("failed to generate upload URL: %w", err)
	}

	return dto.CVUploadURLResponse{
		UploadURL: url,
		CVKey:     key,
	}, nil
}

func sanitiseFileName(name string) string {
	// Remove any directory traversal attempts
	parts := strings.Split(name, "/")
	name = parts[len(parts)-1]
	// Replace spaces with hyphens
	name = strings.ReplaceAll(name, " ", "-")
	return name
}
