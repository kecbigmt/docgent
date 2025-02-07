package workflow

import (
	"context"
	"errors"
	"testing"

	"docgent-backend/internal/domain"
	"docgent-backend/internal/domain/tooluse"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type MockChatModel struct {
	mock.Mock
}

func (m *MockChatModel) SetSystemInstruction(instruction string) error {
	args := m.Called(instruction)
	return args.Error(0)
}

func (m *MockChatModel) SendMessage(ctx context.Context, message domain.Message) (string, error) {
	args := m.Called(ctx, message)
	return args.String(0), args.Error(1)
}

func (m *MockChatModel) GetHistory() ([]domain.Message, error) {
	args := m.Called()
	return args.Get(0).([]domain.Message), args.Error(1)
}

type MockConversationService struct {
	mock.Mock
}

func (m *MockConversationService) Reply(input string) error {
	args := m.Called(input)
	return args.Error(0)
}

type MockFileQueryService struct {
	mock.Mock
}

func (m *MockFileQueryService) FindFile(ctx context.Context, path string) (domain.File, error) {
	args := m.Called(ctx, path)
	return args.Get(0).(domain.File), args.Error(1)
}

type MockFileChangeService struct {
	mock.Mock
}

func (m *MockFileChangeService) CreateFile(ctx context.Context, path, content string) error {
	args := m.Called(ctx, path, content)
	return args.Error(0)
}

func (m *MockFileChangeService) DeleteFile(ctx context.Context, path string) error {
	args := m.Called(ctx, path)
	return args.Error(0)
}

func (m *MockFileChangeService) ModifyFile(ctx context.Context, path string, hunks []tooluse.Hunk) error {
	args := m.Called(ctx, path, hunks)
	return args.Error(0)
}

func (m *MockFileChangeService) RenameFile(ctx context.Context, oldPath, newPath string, hunks []tooluse.Hunk) error {
	args := m.Called(ctx, oldPath, newPath, hunks)
	return args.Error(0)
}

type MockProposalRepository struct {
	mock.Mock
}

func (m *MockProposalRepository) CreateProposal(diffs domain.Diffs, content domain.ProposalContent) (domain.ProposalHandle, error) {
	args := m.Called(diffs, content)
	return args.Get(0).(domain.ProposalHandle), args.Error(1)
}

func (m *MockProposalRepository) GetProposal(handle domain.ProposalHandle) (domain.Proposal, error) {
	args := m.Called(handle)
	return args.Get(0).(domain.Proposal), args.Error(1)
}

func (m *MockProposalRepository) NewProposalHandle(value string) domain.ProposalHandle {
	args := m.Called(value)
	return args.Get(0).(domain.ProposalHandle)
}

func (m *MockProposalRepository) CreateComment(handle domain.ProposalHandle, commentBody string) (domain.Comment, error) {
	args := m.Called(handle, commentBody)
	return args.Get(0).(domain.Comment), args.Error(1)
}

func (m *MockProposalRepository) UpdateProposalContent(handle domain.ProposalHandle, content domain.ProposalContent) error {
	args := m.Called(handle, content)
	return args.Error(0)
}

func (m *MockProposalRepository) ApplyProposalDiffs(handle domain.ProposalHandle, diffs domain.Diffs) error {
	args := m.Called(handle, diffs)
	return args.Error(0)
}

func TestProposalGenerateWorkflow_Execute(t *testing.T) {
	tests := []struct {
		name           string
		chatHistory    []ChatMessage
		setupMocks     func(*MockChatModel, *MockConversationService, *MockFileQueryService, *MockFileChangeService, *MockProposalRepository)
		expectedHandle domain.ProposalHandle
		expectedError  error
	}{
		{
			name: "正常系：提案が正常に生成される",
			chatHistory: []ChatMessage{
				{Author: "user", Content: "APIの仕様書を作成してください"},
				{Author: "assistant", Content: "承知しました。どのような内容を含めるべきでしょうか？"},
				{Author: "user", Content: "エンドポイント、リクエスト、レスポンスの形式を含めてください"},
			},
			setupMocks: func(chatModel *MockChatModel, conversationService *MockConversationService, fileQueryService *MockFileQueryService, fileChangeService *MockFileChangeService, proposalRepository *MockProposalRepository) {
				chatModel.On("SetSystemInstruction", mock.Anything).Return(nil)

				chatModel.On("SendMessage", mock.Anything, mock.Anything).Return(`<create_file><path>path/to/file.md</path><content>Hello, world!</content></create_file>`, nil).Once()
				fileChangeService.On("CreateFile", mock.Anything, "path/to/file.md", "Hello, world!").Return(nil)

				chatModel.On("SendMessage", mock.Anything, mock.Anything).Return(`<create_proposal><title>API仕様書の作成</title><description>APIの仕様書を作成します。エンドポイント、リクエスト、レスポンスの形式を含めます。</description></create_proposal>`, nil).Once()

				proposalHandle := domain.NewProposalHandle("github", "123")
				proposalRepository.On("CreateProposal", domain.Diffs{}, mock.MatchedBy(func(content domain.ProposalContent) bool {
					return content.Title == "API仕様書の作成"
				})).Return(proposalHandle, nil)

				chatModel.On("SendMessage", mock.Anything, mock.Anything).Return(`<attempt_complete><message>提案を作成しました</message></attempt_complete>`, nil).Once()
				conversationService.On("Reply", "提案を作成しました").Return(nil)
			},
			expectedHandle: domain.NewProposalHandle("github", "123"),
			expectedError:  nil,
		},
		{
			name: "エラー系：エージェントの実行に失敗する",
			chatHistory: []ChatMessage{
				{Author: "user", Content: "APIの仕様書を作成してください"},
			},
			setupMocks: func(chatModel *MockChatModel, conversationService *MockConversationService, fileQueryService *MockFileQueryService, fileChangeService *MockFileChangeService, proposalRepository *MockProposalRepository) {
				chatModel.On("SetSystemInstruction", mock.Anything).Return(nil)
				chatModel.On("SendMessage", mock.Anything, mock.Anything).Return("", errors.New("failed to generate response"))
				conversationService.On("Reply", "Something went wrong while generating the proposal").Return(nil)
			},
			expectedHandle: domain.ProposalHandle{},
			expectedError:  errors.New("failed to initiate task loop: failed to generate response: failed to generate response"),
		},
		{
			name: "エラー系：提案の作成に失敗する",
			chatHistory: []ChatMessage{
				{Author: "user", Content: "APIの仕様書を作成してください"},
			},
			setupMocks: func(chatModel *MockChatModel, conversationService *MockConversationService, fileQueryService *MockFileQueryService, fileChangeService *MockFileChangeService, proposalRepository *MockProposalRepository) {
				chatModel.On("SetSystemInstruction", mock.Anything).Return(nil)

				chatModel.On("SendMessage", mock.Anything, mock.Anything).Return(`<create_file><path>path/to/file.md</path><content>Hello, world!</content></create_file>`, nil).Once()
				fileChangeService.On("CreateFile", mock.Anything, "path/to/file.md", "Hello, world!").Return(nil)

				chatModel.On("SendMessage", mock.Anything, mock.Anything).Return(`<create_proposal><title>API仕様書の作成</title><description>APIの仕様書を作成します。</description></create_proposal>`, nil).Once()

				proposalRepository.On("CreateProposal", domain.Diffs{}, mock.Anything).Return(domain.ProposalHandle{}, errors.New("failed to create proposal"))
				conversationService.On("Reply", mock.Anything).Return(nil)
			},
			expectedHandle: domain.ProposalHandle{},
			expectedError:  errors.New("failed to initiate task loop: failed to match tool use: failed to create proposal"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// モックの準備
			chatModel := new(MockChatModel)
			conversationService := new(MockConversationService)
			fileQueryService := new(MockFileQueryService)
			fileChangeService := new(MockFileChangeService)
			proposalRepository := new(MockProposalRepository)

			tt.setupMocks(chatModel, conversationService, fileQueryService, fileChangeService, proposalRepository)

			// ワークフローの作成
			workflow := NewProposalGenerateWorkflow(
				chatModel,
				conversationService,
				fileQueryService,
				fileChangeService,
				proposalRepository,
			)

			// テストの実行
			handle, err := workflow.Execute(context.Background(), tt.chatHistory)

			// アサーション
			if tt.expectedError != nil {
				assert.Error(t, err)
				assert.Equal(t, tt.expectedError.Error(), err.Error())
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedHandle, handle)
			}

			// モックの検証
			chatModel.AssertExpectations(t)
			conversationService.AssertExpectations(t)
			fileQueryService.AssertExpectations(t)
			fileChangeService.AssertExpectations(t)
			proposalRepository.AssertExpectations(t)
		})
	}
}
