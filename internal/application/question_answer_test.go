package application

import (
	"errors"
	"testing"

	"docgent-backend/internal/application/port"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// Reuse existing mock definitions

func TestQuestionAnswerUsecase_Execute(t *testing.T) {
	tests := []struct {
		name          string
		question      string
		setupMocks    func(*MockChatModel, *MockChatSession, *MockConversationService, *MockRAGCorpus)
		expectedError error
	}{
		{
			name:     "success: answer question using RAG",
			question: "How do I use the API?",
			setupMocks: func(chatModel *MockChatModel, chatSession *MockChatSession, conversationService *MockConversationService, ragCorpus *MockRAGCorpus) {
				// Setup RAG query result
				ragCorpus.On("Query", mock.Anything, "How do I use the API?", int32(10), float64(0.5)).Return([]port.RAGDocument{
					{
						Content: "API usage documentation",
						Source:  "docs/api.md",
						Score:   0.85,
					},
				}, nil)

				// Setup chat model
				chatModel.On("StartChat", mock.Anything).Return(chatSession)
				chatSession.On("SendMessage", mock.Anything, "How do I use the API?").Return("Let me explain how to use the API.", nil)

				// Setup conversation service
				conversationService.On("Reply", "Let me explain how to use the API.").Return(nil)
			},
			expectedError: nil,
		},
		{
			name:     "error: RAG query fails",
			question: "How do I use the API?",
			setupMocks: func(chatModel *MockChatModel, chatSession *MockChatSession, conversationService *MockConversationService, ragCorpus *MockRAGCorpus) {
				ragCorpus.On("Query", mock.Anything, "How do I use the API?", int32(10), float64(0.5)).Return([]port.RAGDocument{}, errors.New("failed to query RAG corpus"))
			},
			expectedError: errors.New("failed to query RAG corpus"),
		},
		{
			name:     "error: chat model response fails",
			question: "How do I use the API?",
			setupMocks: func(chatModel *MockChatModel, chatSession *MockChatSession, conversationService *MockConversationService, ragCorpus *MockRAGCorpus) {
				ragCorpus.On("Query", mock.Anything, "How do I use the API?", int32(10), float64(0.5)).Return([]port.RAGDocument{
					{
						Content: "API usage documentation",
						Source:  "docs/api.md",
						Score:   0.85,
					},
				}, nil)

				chatModel.On("StartChat", mock.Anything).Return(chatSession)
				chatSession.On("SendMessage", mock.Anything, "How do I use the API?").Return("", errors.New("failed to generate response"))
			},
			expectedError: errors.New("failed to generate response"),
		},
		{
			name:     "error: conversation service reply fails",
			question: "How do I use the API?",
			setupMocks: func(chatModel *MockChatModel, chatSession *MockChatSession, conversationService *MockConversationService, ragCorpus *MockRAGCorpus) {
				ragCorpus.On("Query", mock.Anything, "How do I use the API?", int32(10), float64(0.5)).Return([]port.RAGDocument{
					{
						Content: "API usage documentation",
						Source:  "docs/api.md",
						Score:   0.85,
					},
				}, nil)

				chatModel.On("StartChat", mock.Anything).Return(chatSession)
				chatSession.On("SendMessage", mock.Anything, "How do I use the API?").Return("Let me explain how to use the API.", nil)

				conversationService.On("Reply", "Let me explain how to use the API.").Return(errors.New("failed to reply"))
			},
			expectedError: errors.New("failed to reply"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Prepare mocks
			chatModel := new(MockChatModel)
			chatSession := new(MockChatSession)
			conversationService := new(MockConversationService)
			ragCorpus := new(MockRAGCorpus)

			tt.setupMocks(chatModel, chatSession, conversationService, ragCorpus)

			// Create usecase
			usecase := NewQuestionAnswerUsecase(
				chatModel,
				ragCorpus,
				conversationService,
			)

			// Execute test
			err := usecase.Execute(tt.question)

			// Assert results
			if tt.expectedError != nil {
				assert.Error(t, err)
				assert.Equal(t, tt.expectedError.Error(), err.Error())
			} else {
				assert.NoError(t, err)
			}

			// Verify mock expectations
			chatModel.AssertExpectations(t)
			chatSession.AssertExpectations(t)
			conversationService.AssertExpectations(t)
			ragCorpus.AssertExpectations(t)
		})
	}
}
