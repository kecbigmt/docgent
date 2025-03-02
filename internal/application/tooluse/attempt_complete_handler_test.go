package tooluse

import (
	"fmt"
	"testing"

	"docgent/internal/application/port"
	"docgent/internal/domain/data"
	"docgent/internal/domain/tooluse"

	"github.com/stretchr/testify/assert"
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

type MockConversationService struct {
	mock.Mock
}

func (m *MockConversationService) Reply(input string, withMention bool) error {
	args := m.Called(input, withMention)
	return args.Error(0)
}

func (m *MockConversationService) GetHistory() (port.ConversationHistory, error) {
	args := m.Called()
	return args.Get(0).(port.ConversationHistory), args.Error(1)
}

func (m *MockConversationService) URI() *data.URI {
	args := m.Called()
	return args.Get(0).(*data.URI)
}

func (m *MockConversationService) MarkEyes() error {
	args := m.Called()
	return args.Error(0)
}

func (m *MockConversationService) RemoveEyes() error {
	args := m.Called()
	return args.Error(0)
}

func TestAttemptCompleteHandler_Handle(t *testing.T) {
	tests := []struct {
		name           string
		toolUse        tooluse.AttemptComplete
		setupMocks     func(*MockConversationService, *MockResponseFormatter)
		expectedResult string
		expectedDone   bool
		expectedError  error
	}{
		{
			name: "Success: Message only",
			toolUse: tooluse.NewAttemptComplete(
				[]tooluse.Message{
					tooluse.NewMessage("Here is the answer:\n- Docgent is a agent that can help you with your documentation.\n- Docgent can create documents based on chat history."),
				},
				[]tooluse.Source{},
			),
			setupMocks: func(conversationService *MockConversationService, responseFormatter *MockResponseFormatter) {
				expectedMessage := "Here is the answer:\n- Docgent is a agent that can help you with your documentation.\n- Docgent can create documents based on chat history."
				responseFormatter.On("FormatResponse", mock.Anything).Return(expectedMessage, nil)
				conversationService.On("Reply", expectedMessage, true).Return(nil)
			},
			expectedResult: "",
			expectedDone:   true,
			expectedError:  nil,
		},
		{
			name: "Success: Message with sources",
			toolUse: tooluse.NewAttemptComplete(
				[]tooluse.Message{
					tooluse.NewMessage("Here is the answer:"),
					tooluse.NewMessageWithSourceIDs("- Docgent is a agent that can help you with your documentation", []string{"1", "2"}),
					tooluse.NewMessageWithSourceID("- Docgent can create documents based on chat history.", "2"),
				},
				[]tooluse.Source{
					tooluse.NewSource("1", "https://github.com/owner/repo/blob/a1b2c3d4e5f6g7h8i9j0k1l2m3n4o5p6q7r8s9t0/docs/what-is-docgent.md", "What is Docgent?"),
					tooluse.NewSource("2", "https://github.com/owner/repo/blob/a1b2c3d4e5f6g7h8i9j0k1l2m3n4o5p6q7r8s9t0/docs/docgent-features.md", "Docgent Features"),
				},
			),
			setupMocks: func(conversationService *MockConversationService, responseFormatter *MockResponseFormatter) {
				expectedMessage := "Here is the answer:\n- Docgent is a agent that can help you with your documentation[^1][^2]\n- Docgent can create documents based on chat history.[^2]\n\n[^1]: https://github.com/owner/repo/blob/a1b2c3d4e5f6g7h8i9j0k1l2m3n4o5p6q7r8s9t0/docs/what-is-docgent.md\n[^2]: https://github.com/owner/repo/blob/a1b2c3d4e5f6g7h8i9j0k1l2m3n4o5p6q7r8s9t0/docs/docgent-features.md"
				responseFormatter.On("FormatResponse", mock.Anything).Return(expectedMessage, nil)
				conversationService.On("Reply", expectedMessage, true).Return(nil)
			},
			expectedResult: "",
			expectedDone:   true,
			expectedError:  nil,
		},
		{
			name: "Error: Reply fails",
			toolUse: tooluse.NewAttemptComplete(
				[]tooluse.Message{
					tooluse.NewMessage("Here is the answer:\n- Docgent is a agent that can help you with your documentation.\n- Docgent can create documents based on chat history."),
				},
				[]tooluse.Source{},
			),
			setupMocks: func(conversationService *MockConversationService, responseFormatter *MockResponseFormatter) {
				expectedMessage := "Here is the answer:\n- Docgent is a agent that can help you with your documentation.\n- Docgent can create documents based on chat history."
				responseFormatter.On("FormatResponse", mock.Anything).Return(expectedMessage, nil)
				conversationService.On("Reply", expectedMessage, true).Return(fmt.Errorf("reply error"))
			},
			expectedResult: "",
			expectedDone:   false,
			expectedError:  fmt.Errorf("failed to reply: reply error"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup mocks
			conversationService := new(MockConversationService)
			responseFormatter := new(MockResponseFormatter)
			tt.setupMocks(conversationService, responseFormatter)

			// Create handler
			handler := NewAttemptCompleteHandler(conversationService, responseFormatter)

			// Execute test
			result, done, err := handler.Handle(tt.toolUse)

			// Assertions
			if tt.expectedError != nil {
				assert.Error(t, err)
				assert.Equal(t, tt.expectedError.Error(), err.Error())
			} else {
				assert.NoError(t, err)
			}
			assert.Equal(t, tt.expectedResult, result)
			assert.Equal(t, tt.expectedDone, done)

			// Verify mocks
			conversationService.AssertExpectations(t)
			responseFormatter.AssertExpectations(t)

			// Verify exact arguments passed to FormatResponse
			responseFormatter.AssertCalled(t, "FormatResponse", tt.toolUse)
		})
	}
}
