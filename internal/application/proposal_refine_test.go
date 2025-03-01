package application

import (
	"errors"
	"strings"
	"sync"
	"testing"

	"docgent/internal/application/port"
	"docgent/internal/domain"
	"docgent/internal/domain/data"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestProposalRefineUsecase_Refine(t *testing.T) {
	tests := []struct {
		name           string
		proposalHandle domain.ProposalHandle
		userFeedback   string
		setupMocks     func(*MockChatModel, *MockChatSession, *MockConversationService, *MockFileQueryService, *MockFileRepository, *MockProposalRepository, *MockRAGCorpus, *MockResponseFormatter)
		expectedError  error
	}{
		{
			name:           "正常系：RAGを使用して提案が正常に更新される",
			proposalHandle: domain.NewProposalHandle("github", "123"),
			userFeedback:   "エンドポイントの説明をもう少し詳しくしてください",
			setupMocks: func(chatModel *MockChatModel, chatSession *MockChatSession, conversationService *MockConversationService, fileQueryService *MockFileQueryService, fileRepository *MockFileRepository, proposalRepository *MockProposalRepository, ragCorpus *MockRAGCorpus, responseFormatter *MockResponseFormatter) {
				conversationService.On("MarkEyes").Return(nil).Once()
				conversationService.On("RemoveEyes").Return(nil).Once()
				conversationService.On("URI").Return(data.NewURIUnsafe("https://github.com/123/456/pull/123")).Once()

				proposal := domain.Proposal{
					Handle: domain.NewProposalHandle("github", "123"),
					Diffs: domain.Diffs{
						{NewName: "docs/api.md"},
					},
				}
				proposalRepository.On("GetProposal", proposal.Handle).Return(proposal, nil)

				fileQueryService.On("GetTree", mock.Anything, mock.AnythingOfType("[]port.GetTreeOption")).Return([]port.TreeMetadata{
					{Path: "docs/api.md", Type: port.NodeTypeFile, Size: 100},
				}, nil)

				chatModel.On("StartChat", mock.Anything).Return(chatSession)

				// 1回目のメッセージ：RAGクエリを実行
				chatSession.On("SendMessage", mock.Anything, mock.Anything).Return(`<query_rag><query>APIドキュメント エンドポイント 詳細</query></query_rag>`, nil).Once()
				// RAGクエリの結果を設定
				ragCorpus.On("Query", mock.Anything, "APIドキュメント エンドポイント 詳細", int32(10), float64(0.7)).Return([]port.RAGDocument{
					{
						Content: "既存のAPIドキュメント",
						Source:  "docs/api.md",
						Score:   0.85,
					},
				}, nil).Once()

				// 2回目のメッセージ：ファイルを更新
				chatSession.On("SendMessage", mock.Anything, mock.Anything).Return(`<modify_file><path>docs/api.md</path><hunk><search>エンドポイント</search><replace>endpoint</replace></hunk></modify_file>`, nil).Once()
				fileRepository.On("Get", mock.Anything, "docs/api.md").Return(&data.File{
					Path:    "docs/api.md",
					Content: "エンドポイントの説明",
				}, nil)
				fileRepository.On("Update", mock.Anything, mock.MatchedBy(func(file *data.File) bool {
					return file.Path == "docs/api.md" && strings.Contains(file.Content, "endpoint")
				})).Return(nil)

				// 3回目のメッセージ：タスクを完了
				chatSession.On("SendMessage", mock.Anything, mock.Anything).Return(`<attempt_complete><message>提案を更新しました</message></attempt_complete>`, nil).Once()
				responseFormatter.On("FormatResponse", mock.Anything).Return("提案を更新しました", nil).Once()
				conversationService.On("Reply", "提案を更新しました", true).Return(nil)

			},
			expectedError: nil,
		},
		{
			name:           "エラー系：エージェントの実行に失敗する",
			proposalHandle: domain.NewProposalHandle("github", "123"),
			userFeedback:   "エンドポイントの説明をもう少し詳しくしてください",
			setupMocks: func(chatModel *MockChatModel, chatSession *MockChatSession, conversationService *MockConversationService, fileQueryService *MockFileQueryService, fileRepository *MockFileRepository, proposalRepository *MockProposalRepository, ragCorpus *MockRAGCorpus, responseFormatter *MockResponseFormatter) {
				conversationService.On("MarkEyes").Return(nil).Once()
				conversationService.On("RemoveEyes").Return(nil).Once()
				conversationService.On("URI").Return(data.NewURIUnsafe("https://github.com/123/456/pull/123")).Once()

				proposal := domain.Proposal{
					Handle: domain.NewProposalHandle("github", "123"),
					Diffs: domain.Diffs{
						{NewName: "docs/api.md"},
					},
				}
				proposalRepository.On("GetProposal", proposal.Handle).Return(proposal, nil)

				fileQueryService.On("GetTree", mock.Anything, mock.AnythingOfType("[]port.GetTreeOption")).Return([]port.TreeMetadata{
					{Path: "docs/api.md", Type: port.NodeTypeFile, Size: 100},
				}, nil)

				chatModel.On("StartChat", mock.Anything).Return(chatSession)
				chatSession.On("SendMessage", mock.Anything, mock.Anything).Return("", errors.New("failed to generate response"))
				conversationService.On("Reply", "Something went wrong while refining the proposal", true).Return(nil)
			},
			expectedError: errors.New("failed to initiate task loop: failed to generate response: failed to generate response"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// モックの準備
			chatModel := new(MockChatModel)
			chatSession := new(MockChatSession)
			conversationService := new(MockConversationService)
			conversationService.markEyesWaitGroup = &sync.WaitGroup{}
			conversationService.markEyesWaitGroup.Add(1)
			fileQueryService := new(MockFileQueryService)
			fileRepository := new(MockFileRepository)
			proposalRepository := new(MockProposalRepository)
			ragCorpus := new(MockRAGCorpus)
			responseFormatter := new(MockResponseFormatter)

			tt.setupMocks(chatModel, chatSession, conversationService, fileQueryService, fileRepository, proposalRepository, ragCorpus, responseFormatter)

			// ワークフローの作成
			workflow := NewProposalRefineUsecase(
				chatModel,
				conversationService,
				fileQueryService,
				fileRepository,
				[]port.SourceRepository{},
				proposalRepository,
				responseFormatter,
				WithProposalRefineRAGCorpus(ragCorpus),
			)

			// テストの実行
			err := workflow.Refine(tt.proposalHandle, tt.userFeedback)

			conversationService.markEyesWaitGroup.Wait()

			// アサーション
			if tt.expectedError != nil {
				assert.Error(t, err)
				assert.Equal(t, tt.expectedError.Error(), err.Error())
			} else {
				assert.NoError(t, err)
			}

			// モックの検証
			chatModel.AssertExpectations(t)
			conversationService.AssertExpectations(t)
			fileQueryService.AssertExpectations(t)
			fileRepository.AssertExpectations(t)
			proposalRepository.AssertExpectations(t)
			ragCorpus.AssertExpectations(t)
		})
	}
}
