package application

import (
	"errors"
	"testing"

	"docgent-backend/internal/application/port"
	"docgent-backend/internal/domain"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestProposalRefineUsecase_Refine(t *testing.T) {
	tests := []struct {
		name           string
		proposalHandle domain.ProposalHandle
		userFeedback   string
		setupMocks     func(*MockChatModel, *MockChatSession, *MockConversationService, *MockFileQueryService, *MockFileChangeService, *MockProposalRepository, *MockRAGCorpus)
		expectedError  error
	}{
		{
			name:           "正常系：RAGを使用して提案が正常に更新される",
			proposalHandle: domain.NewProposalHandle("github", "123"),
			userFeedback:   "エンドポイントの説明をもう少し詳しくしてください",
			setupMocks: func(chatModel *MockChatModel, chatSession *MockChatSession, conversationService *MockConversationService, fileQueryService *MockFileQueryService, fileChangeService *MockFileChangeService, proposalRepository *MockProposalRepository, ragCorpus *MockRAGCorpus) {
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
				fileChangeService.On("ModifyFile", mock.Anything, "docs/api.md", mock.Anything).Return(nil)

				// 3回目のメッセージ：タスクを完了
				chatSession.On("SendMessage", mock.Anything, mock.Anything).Return(`<attempt_complete><message>提案を更新しました</message></attempt_complete>`, nil).Once()
				conversationService.On("Reply", "提案を更新しました").Return(nil)
			},
			expectedError: nil,
		},
		{
			name:           "エラー系：エージェントの実行に失敗する",
			proposalHandle: domain.NewProposalHandle("github", "123"),
			userFeedback:   "エンドポイントの説明をもう少し詳しくしてください",
			setupMocks: func(chatModel *MockChatModel, chatSession *MockChatSession, conversationService *MockConversationService, fileQueryService *MockFileQueryService, fileChangeService *MockFileChangeService, proposalRepository *MockProposalRepository, ragCorpus *MockRAGCorpus) {
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
				conversationService.On("Reply", "Something went wrong while refining the proposal").Return(nil)
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
			fileQueryService := new(MockFileQueryService)
			fileChangeService := new(MockFileChangeService)
			proposalRepository := new(MockProposalRepository)
			ragCorpus := new(MockRAGCorpus)

			tt.setupMocks(chatModel, chatSession, conversationService, fileQueryService, fileChangeService, proposalRepository, ragCorpus)

			// ワークフローの作成
			workflow := NewProposalRefineUsecase(
				chatModel,
				conversationService,
				fileQueryService,
				fileChangeService,
				proposalRepository,
				ragCorpus,
			)

			// テストの実行
			err := workflow.Refine(tt.proposalHandle, tt.userFeedback)

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
			fileChangeService.AssertExpectations(t)
			proposalRepository.AssertExpectations(t)
			ragCorpus.AssertExpectations(t)
		})
	}
}
