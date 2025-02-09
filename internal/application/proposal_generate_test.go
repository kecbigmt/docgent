package application

import (
	"context"
	"errors"
	"io"
	"testing"

	"docgent-backend/internal/application/port"
	"docgent-backend/internal/domain"
	"docgent-backend/internal/domain/tooluse"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type MockChatModel struct {
	mock.Mock
}

func (m *MockChatModel) StartChat(systemInstruction string) domain.ChatSession {
	args := m.Called(systemInstruction)
	return args.Get(0).(domain.ChatSession)
}

type MockChatSession struct {
	mock.Mock
}

func (m *MockChatSession) SendMessage(ctx context.Context, message string) (string, error) {
	args := m.Called(ctx, message)
	return args.String(0), args.Error(1)
}

func (m *MockChatSession) GetHistory() ([]domain.Message, error) {
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

func (m *MockConversationService) GetHistory() ([]port.ConversationMessage, error) {
	args := m.Called()
	return args.Get(0).([]port.ConversationMessage), args.Error(1)
}

func (m *MockConversationService) MarkEyes() error {
	args := m.Called()
	return args.Error(0)
}

func (m *MockConversationService) RemoveEyes() error {
	args := m.Called()
	return args.Error(0)
}

type MockFileQueryService struct {
	mock.Mock
}

func (m *MockFileQueryService) FindFile(ctx context.Context, path string) (port.File, error) {
	args := m.Called(ctx, path)
	return args.Get(0).(port.File), args.Error(1)
}

func (m *MockFileQueryService) GetTree(ctx context.Context, options ...port.GetTreeOption) ([]port.TreeMetadata, error) {
	args := m.Called(ctx, options)
	return args.Get(0).([]port.TreeMetadata), args.Error(1)
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

type MockRAGCorpus struct {
	mock.Mock
}

func (m *MockRAGCorpus) Query(ctx context.Context, query string, similarityTopK int32, vectorDistanceThreshold float64) ([]port.RAGDocument, error) {
	args := m.Called(ctx, query, similarityTopK, vectorDistanceThreshold)
	return args.Get(0).([]port.RAGDocument), args.Error(1)
}

func (m *MockRAGCorpus) UploadFile(ctx context.Context, file io.Reader, fileName string, options ...port.RAGCorpusUploadFileOption) error {
	args := m.Called(ctx, file, fileName, options)
	return args.Error(0)
}

func (m *MockRAGCorpus) ListFiles(ctx context.Context) ([]port.RAGFile, error) {
	args := m.Called(ctx)
	return args.Get(0).([]port.RAGFile), args.Error(1)
}

func (m *MockRAGCorpus) DeleteFile(ctx context.Context, fileID int64) error {
	args := m.Called(ctx, fileID)
	return args.Error(0)
}

func TestProposalGenerateUsecase_Execute(t *testing.T) {
	tests := []struct {
		name           string
		setupMocks     func(*MockChatModel, *MockChatSession, *MockConversationService, *MockFileQueryService, *MockFileChangeService, *MockProposalRepository, *MockRAGCorpus)
		expectedHandle domain.ProposalHandle
		expectedError  error
	}{
		{
			name: "正常系：RAGを使用して提案が正常に生成される",
			setupMocks: func(chatModel *MockChatModel, chatSession *MockChatSession, conversationService *MockConversationService, fileQueryService *MockFileQueryService, fileChangeService *MockFileChangeService, proposalRepository *MockProposalRepository, ragCorpus *MockRAGCorpus) {
				conversationService.On("MarkEyes").Return(nil)
				conversationService.On("RemoveEyes").Return(nil)

				conversationService.On("GetHistory").Return([]port.ConversationMessage{
					{Author: "user", Content: "APIの仕様書を作成してください"},
					{Author: "assistant", Content: "承知しました。どのような内容を含めるべきでしょうか？"},
					{Author: "user", Content: "エンドポイント、リクエスト、レスポンスの形式を含めてください"},
				}, nil)

				fileQueryService.On("GetTree", mock.Anything, mock.AnythingOfType("[]port.GetTreeOption")).Return([]port.TreeMetadata{
					{Path: "docs/api.md", Type: port.NodeTypeFile, Size: 100},
				}, nil)

				chatModel.On("StartChat", mock.Anything).Return(chatSession)

				// 1回目のメッセージ：RAGクエリを実行
				chatSession.On("SendMessage", mock.Anything, mock.Anything).Return(`<query_rag><query>APIドキュメント 仕様書 エンドポイント</query></query_rag>`, nil).Once()
				// RAGクエリの結果を設定
				ragCorpus.On("Query", mock.Anything, "APIドキュメント 仕様書 エンドポイント", int32(10), float64(0.7)).Return([]port.RAGDocument{
					{
						Content: "既存のAPIドキュメント",
						Source:  "docs/api.md",
						Score:   0.85,
					},
				}, nil).Once()

				// 2回目のメッセージ：ファイルを作成
				chatSession.On("SendMessage", mock.Anything, mock.Anything).Return(`<create_file><path>path/to/file.md</path><content>Hello, world!</content></create_file>`, nil).Once()
				fileChangeService.On("CreateFile", mock.Anything, "path/to/file.md", "Hello, world!").Return(nil)

				// 3回目のメッセージ：提案を作成
				chatSession.On("SendMessage", mock.Anything, mock.Anything).Return(`<create_proposal><title>API仕様書の作成</title><description>APIの仕様書を作成します。エンドポイント、リクエスト、レスポンスの形式を含めます。</description></create_proposal>`, nil).Once()
				proposalHandle := domain.NewProposalHandle("github", "123")
				proposalRepository.On("CreateProposal", domain.Diffs{}, mock.MatchedBy(func(content domain.ProposalContent) bool {
					return content.Title == "API仕様書の作成"
				})).Return(proposalHandle, nil)

				// 4回目のメッセージ：タスクを完了
				chatSession.On("SendMessage", mock.Anything, mock.Anything).Return(`<attempt_complete><message>提案を作成しました</message></attempt_complete>`, nil).Once()
				conversationService.On("Reply", "提案を作成しました").Return(nil)
			},
			expectedHandle: domain.NewProposalHandle("github", "123"),
			expectedError:  nil,
		},
		{
			name: "エラー系：エージェントの実行に失敗する",
			setupMocks: func(chatModel *MockChatModel, chatSession *MockChatSession, conversationService *MockConversationService, fileQueryService *MockFileQueryService, fileChangeService *MockFileChangeService, proposalRepository *MockProposalRepository, ragCorpus *MockRAGCorpus) {
				conversationService.On("MarkEyes").Return(nil)
				conversationService.On("RemoveEyes").Return(nil)
				conversationService.On("GetHistory").Return([]port.ConversationMessage{
					{Author: "user", Content: "APIの仕様書を作成してください"},
					{Author: "assistant", Content: "承知しました。どのような内容を含めるべきでしょうか？"},
					{Author: "user", Content: "エンドポイント、リクエスト、レスポンスの形式を含めてください"},
				}, nil)
				fileQueryService.On("GetTree", mock.Anything, mock.AnythingOfType("[]port.GetTreeOption")).Return([]port.TreeMetadata{
					{Path: "docs/api.md", Type: port.NodeTypeFile, Size: 100},
				}, nil)
				chatModel.On("StartChat", mock.Anything).Return(chatSession)
				chatSession.On("SendMessage", mock.Anything, mock.Anything).Return("", errors.New("failed to generate response"))
				conversationService.On("Reply", "Something went wrong while generating the proposal").Return(nil)
			},
			expectedHandle: domain.ProposalHandle{},
			expectedError:  errors.New("failed to initiate task loop: failed to generate response: failed to generate response"),
		},
		{
			name: "エラー系：提案の作成に失敗する",
			setupMocks: func(chatModel *MockChatModel, chatSession *MockChatSession, conversationService *MockConversationService, fileQueryService *MockFileQueryService, fileChangeService *MockFileChangeService, proposalRepository *MockProposalRepository, ragCorpus *MockRAGCorpus) {
				conversationService.On("MarkEyes").Return(nil)
				conversationService.On("RemoveEyes").Return(nil)

				conversationService.On("GetHistory").Return([]port.ConversationMessage{
					{Author: "user", Content: "APIの仕様書を作成してください"},
					{Author: "assistant", Content: "承知しました。どのような内容を含めるべきでしょうか？"},
					{Author: "user", Content: "エンドポイント、リクエスト、レスポンスの形式を含めてください"},
				}, nil)
				fileQueryService.On("GetTree", mock.Anything, mock.AnythingOfType("[]port.GetTreeOption")).Return([]port.TreeMetadata{
					{Path: "docs/api.md", Type: port.NodeTypeFile, Size: 100},
				}, nil)

				chatModel.On("StartChat", mock.Anything).Return(chatSession)

				chatSession.On("SendMessage", mock.Anything, mock.Anything).Return(`<create_file><path>path/to/file.md</path><content>Hello, world!</content></create_file>`, nil).Once()
				fileChangeService.On("CreateFile", mock.Anything, "path/to/file.md", "Hello, world!").Return(nil)

				chatSession.On("SendMessage", mock.Anything, mock.Anything).Return(`<create_proposal><title>API仕様書の作成</title><description>APIの仕様書を作成します。</description></create_proposal>`, nil).Once()

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
			chatSession := new(MockChatSession)
			conversationService := new(MockConversationService)
			fileQueryService := new(MockFileQueryService)
			fileChangeService := new(MockFileChangeService)
			proposalRepository := new(MockProposalRepository)
			ragCorpus := new(MockRAGCorpus)

			tt.setupMocks(chatModel, chatSession, conversationService, fileQueryService, fileChangeService, proposalRepository, ragCorpus)

			// ワークフローの作成
			workflow := NewProposalGenerateUsecase(
				chatModel,
				conversationService,
				fileQueryService,
				fileChangeService,
				proposalRepository,
				ragCorpus,
			)

			// テストの実行
			handle, err := workflow.Execute(context.Background())

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
			ragCorpus.AssertExpectations(t)
		})
	}
}
