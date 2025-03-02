package application

import (
	"context"

	"docgent/internal/application/port"
	"docgent/internal/domain/data"

	"github.com/stretchr/testify/mock"
)

// MockFileQueryService is a mock implementation of the port.FileQueryService interface
type MockFileQueryService struct {
	mock.Mock
}

func (m *MockFileQueryService) FindFile(ctx context.Context, path string) (data.File, error) {
	args := m.Called(ctx, path)
	return args.Get(0).(data.File), args.Error(1)
}

func (m *MockFileQueryService) GetTree(ctx context.Context, options ...port.GetTreeOption) ([]port.TreeMetadata, error) {
	args := m.Called(ctx, options)
	return args.Get(0).([]port.TreeMetadata), args.Error(1)
}

func (m *MockFileQueryService) GetURI(ctx context.Context, path string) (*data.URI, error) {
	args := m.Called(ctx, path)
	return data.NewURIUnsafe(args.Get(0).(string)), args.Error(1)
}

func (m *MockFileQueryService) GetFilePath(uri *data.URI) (string, error) {
	args := m.Called(uri)
	return args.String(0), args.Error(1)
}
