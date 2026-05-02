package mocks

import (
	"context"

	"github.com/stretchr/testify/mock"
)

type MockCVStorage struct {
	mock.Mock
}

func (m *MockCVStorage) GenerateUploadURL(ctx context.Context, key string) (string, error) {
	args := m.Called(ctx, key)
	return args.String(0), args.Error(1)
}

func (m *MockCVStorage) GenerateDownloadURL(ctx context.Context, key string) (string, error) {
	args := m.Called(ctx, key)
	return args.String(0), args.Error(1)
}

func (m *MockCVStorage) DeleteObject(ctx context.Context, key string) error {
	args := m.Called(ctx, key)
	return args.Error(0)
}
