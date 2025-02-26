package application

import (
	"errors"
	"sync"
	"testing"

	"docgent/internal/application/port"

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
		disableRAG    bool
	}{
		{
			name:     "success: answer question using RAG",
			question: "How do I use the API?",
			setupMocks: func(chatModel *MockChatModel, chatSession *MockChatSession, conversationService *MockConversationService, ragCorpus *MockRAGCorpus) {
				conversationService.On("MarkEyes").Return(nil).Once()
				conversationService.On("RemoveEyes").Return(nil).Once()

				// RAG query
				ragCorpus.On("Query", mock.Anything, "How do I use the API?", int32(10), float64(0.5)).Return([]port.RAGDocument{
					{
						Content: "API usage documentation",
						Source:  "docs/api.md",
						Score:   0.85,
					},
				}, nil)

				// Chat model
				chatModel.On("StartChat", mock.Anything).Return(chatSession)
				chatSession.On("SendMessage", mock.Anything, "How do I use the API?").Return("Let me explain how to use the API.", nil)

				// Reply
				conversationService.On("Reply", "Let me explain how to use the API.", false).Return(nil)
			},
			expectedError: nil,
		},
		{
			name:     "success: answer question without RAG",
			question: "How do I use the API?",
			setupMocks: func(chatModel *MockChatModel, chatSession *MockChatSession, conversationService *MockConversationService, ragCorpus *MockRAGCorpus) {
				conversationService.On("MarkEyes").Return(nil).Once()
				conversationService.On("RemoveEyes").Return(nil).Once()

				chatModel.On("StartChat", "You are a helpful assistant. Unfortunately, you do not have access to any domain-specific knowledge. Answer the question based on the general knowledge.").Return(chatSession)
				chatSession.On("SendMessage", mock.Anything, "How do I use the API?").Return("Let me explain how to use the API.", nil)
				conversationService.On("Reply", "Let me explain how to use the API.", false).Return(nil)
			},
			expectedError: nil,
			disableRAG:    true,
		},
		{
			name:     "error: RAG query fails",
			question: "How do I use the API?",
			setupMocks: func(chatModel *MockChatModel, chatSession *MockChatSession, conversationService *MockConversationService, ragCorpus *MockRAGCorpus) {
				conversationService.On("MarkEyes").Return(nil).Once()
				conversationService.On("RemoveEyes").Return(nil).Once()
				ragCorpus.On("Query", mock.Anything, "How do I use the API?", int32(10), float64(0.5)).Return([]port.RAGDocument{}, errors.New("failed to query RAG corpus"))
			},
			expectedError: errors.New("failed to query RAG corpus"),
		},
		{
			name:     "error: chat model response fails",
			question: "How do I use the API?",
			setupMocks: func(chatModel *MockChatModel, chatSession *MockChatSession, conversationService *MockConversationService, ragCorpus *MockRAGCorpus) {
				conversationService.On("MarkEyes").Return(nil).Once()
				conversationService.On("RemoveEyes").Return(nil).Once()

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
				conversationService.On("MarkEyes").Return(nil).Once()
				conversationService.On("RemoveEyes").Return(nil).Once()
				conversationService.On("Reply", "Let me explain how to use the API.", false).Return(errors.New("failed to reply"))

				ragCorpus.On("Query", mock.Anything, "How do I use the API?", int32(10), float64(0.5)).Return([]port.RAGDocument{
					{
						Content: "API usage documentation",
						Source:  "docs/api.md",
						Score:   0.85,
					},
				}, nil)

				chatModel.On("StartChat", mock.Anything).Return(chatSession)
				chatSession.On("SendMessage", mock.Anything, "How do I use the API?").Return("Let me explain how to use the API.", nil)
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
			conversationService.markEyesWaitGroup = &sync.WaitGroup{}
			conversationService.markEyesWaitGroup.Add(1)
			ragCorpus := new(MockRAGCorpus)

			tt.setupMocks(chatModel, chatSession, conversationService, ragCorpus)

			var options []NewQuestionAnswerUsecaseOption
			if !tt.disableRAG {
				options = append(options, WithQuestionAnswerRAGCorpus(ragCorpus))
			}

			// Create usecase
			usecase := NewQuestionAnswerUsecase(
				chatModel,
				conversationService,
				options...,
			)

			// Execute test
			err := usecase.Execute(tt.question)

			conversationService.markEyesWaitGroup.Wait()

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
