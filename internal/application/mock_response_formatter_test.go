package application

import (
	"docgent/internal/domain/tooluse"

	"github.com/stretchr/testify/mock"
)

// MockResponseFormatter is a mock implementation of the ResponseFormatter interface
type MockResponseFormatter struct {
	mock.Mock
}

func (m *MockResponseFormatter) FormatResponse(toolUse tooluse.AttemptComplete) (string, error) {
	args := m.Called(toolUse)
	return args.String(0), args.Error(1)
}
