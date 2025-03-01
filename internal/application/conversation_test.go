package application

import (
	"context"
	"errors"
	"sync"
	"testing"

	"docgent/internal/application/port"
	"docgent/internal/domain/data"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockSourceRepositoryがport.SourceRepositoryインターフェースを実装するよう修正
type MockSourceRepository struct {
	mock.Mock
}

func (m *MockSourceRepository) Match(uri *data.URI) bool {
	args := m.Called(uri)
	return args.Bool(0)
}

func (m *MockSourceRepository) Find(ctx context.Context, uri *data.URI) (*data.Source, error) {
	args := m.Called(ctx, uri)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*data.Source), args.Error(1)
}

func TestConversationUsecase_Execute(t *testing.T) {
	tests := []struct {
		name          string
		setupMocks    func(*MockChatModel, *MockChatSession, *MockConversationService, *MockFileQueryService, *MockSourceRepository, *MockRAGCorpus, *MockResponseFormatter)
		expectedError error
		disableRAG    bool
	}{
		{
			name: "正常系：基本的な会話が成功する",
			setupMocks: func(chatModel *MockChatModel, chatSession *MockChatSession, conversationService *MockConversationService, fileQueryService *MockFileQueryService, sourceRepository *MockSourceRepository, ragCorpus *MockRAGCorpus, responseFormatter *MockResponseFormatter) {
				conversationService.On("MarkEyes").Return(nil).Once()
				conversationService.On("RemoveEyes").Return(nil).Once()

				// 会話履歴を返す
				conversationService.On("GetHistory").Return(port.ConversationHistory{
					URI: data.NewURIUnsafe("https://app.slack.com/client/T00000000/C00000000/thread/T00000000-00000000"),
					Messages: []port.ConversationMessage{
						{Author: "user", Content: "こんにちは"},
						{Author: "docgent", Content: "こんにちは！何かお手伝いできることはありますか？", IsYou: true},
						{Author: "user", Content: "APIの使い方を教えてください", YouMentioned: true},
					},
				}, nil).Once()

				// チャットモデルの設定
				chatModel.On("StartChat", mock.Anything).Return(chatSession).Once()

				// エージェントのタスクループの実行
				// 1回目のメッセージ：RAGクエリを実行
				chatSession.On("SendMessage", mock.Anything, mock.Anything).Return(`<query_rag><query>APIの使い方</query></query_rag>`, nil).Once()

				// RAGクエリの結果
				ragCorpus.On("Query", mock.Anything, "APIの使い方", int32(10), float64(0.7)).Return([]port.RAGDocument{
					{
						Content: "APIの使い方ドキュメント",
						Source:  "docs/api.md",
						Score:   0.85,
					},
				}, nil).Once()

				// 2回目のメッセージ：ファイル内容を確認
				chatSession.On("SendMessage", mock.Anything, mock.Anything).Return(`<find_file><path>docs/api.md</path></find_file>`, nil).Once()

				// ファイル内容を返す
				fileQueryService.On("FindFile", mock.Anything, "docs/api.md").Return(data.File{
					Path:    "docs/api.md",
					Content: "# API使用方法\n\nこのAPIは以下のエンドポイントを提供しています...",
				}, nil).Once()

				// 3回目のメッセージ：解答を生成
				chatSession.On("SendMessage", mock.Anything, mock.Anything).Return(`<attempt_complete>
APIの使い方について説明します。ドキュメントによると、このAPIは複数のエンドポイントを提供しており...
</attempt_complete>`, nil).Once()

				// ResponseFormatterのモック設定
				responseFormatter.On("FormatResponse", mock.Anything).Return("APIの使い方について説明します。ドキュメントによると、このAPIは複数のエンドポイントを提供しており...", nil).Once()

				// 回答を返す
				conversationService.On("Reply", mock.Anything, true).Return(nil).Once()
			},
			expectedError: nil,
		},
		{
			name: "正常系：RAGなしで会話が成功する",
			setupMocks: func(chatModel *MockChatModel, chatSession *MockChatSession, conversationService *MockConversationService, fileQueryService *MockFileQueryService, sourceRepository *MockSourceRepository, ragCorpus *MockRAGCorpus, responseFormatter *MockResponseFormatter) {
				conversationService.On("MarkEyes").Return(nil).Once()
				conversationService.On("RemoveEyes").Return(nil).Once()

				// 会話履歴を返す
				conversationService.On("GetHistory").Return(port.ConversationHistory{
					URI: data.NewURIUnsafe("https://app.slack.com/client/T00000000/C00000000/thread/T00000000-00000000"),
					Messages: []port.ConversationMessage{
						{Author: "user", Content: "こんにちは"},
						{Author: "docgent", Content: "こんにちは！何かお手伝いできることはありますか？", IsYou: true},
						{Author: "user", Content: "調子はどう？", YouMentioned: true},
					},
				}, nil).Once()

				// チャットモデルの設定
				chatModel.On("StartChat", mock.Anything).Return(chatSession).Once()

				// メッセージの送信
				chatSession.On("SendMessage", mock.Anything, mock.Anything).Return(`<attempt_complete>
こんにちは！私は元気です。あなたはどうですか？
</attempt_complete>`, nil).Once()

				// ResponseFormatterのモック設定
				responseFormatter.On("FormatResponse", mock.Anything).Return("こんにちは！私は元気です。あなたはどうですか？", nil).Once()

				// 回答を返す
				conversationService.On("Reply", mock.Anything, true).Return(nil).Once()
			},
			expectedError: nil,
			disableRAG:    true,
		},
		{
			name: "エラー系：会話履歴の取得に失敗する",
			setupMocks: func(chatModel *MockChatModel, chatSession *MockChatSession, conversationService *MockConversationService, fileQueryService *MockFileQueryService, sourceRepository *MockSourceRepository, ragCorpus *MockRAGCorpus, responseFormatter *MockResponseFormatter) {
				conversationService.On("MarkEyes").Return(nil).Once()
				conversationService.On("RemoveEyes").Return(nil).Once()

				// 会話履歴の取得に失敗
				conversationService.On("GetHistory").Return(port.ConversationHistory{}, errors.New("failed to retrieve conversation history")).Once()
			},
			expectedError: errors.New("failed to get chat history"),
		},
		{
			name: "エラー系：タスク実行ループでエラーが発生する",
			setupMocks: func(chatModel *MockChatModel, chatSession *MockChatSession, conversationService *MockConversationService, fileQueryService *MockFileQueryService, sourceRepository *MockSourceRepository, ragCorpus *MockRAGCorpus, responseFormatter *MockResponseFormatter) {
				conversationService.On("MarkEyes").Return(nil).Once()
				conversationService.On("RemoveEyes").Return(nil).Once()

				// 会話履歴を返す
				conversationService.On("GetHistory").Return(port.ConversationHistory{
					URI: data.NewURIUnsafe("https://app.slack.com/client/T00000000/C00000000/thread/T00000000-00000000"),
					Messages: []port.ConversationMessage{
						{Author: "user", Content: "こんにちは", YouMentioned: true},
					},
				}, nil).Once()

				// チャットモデルの設定
				chatModel.On("StartChat", mock.Anything).Return(chatSession).Once()

				// メッセージ送信でエラー
				chatSession.On("SendMessage", mock.Anything, mock.Anything).Return("", errors.New("failed to generate response")).Once()

				// エラー通知
				conversationService.On("Reply", "Something went wrong. Please try again later.", true).Return(nil).Once()
			},
			expectedError: errors.New("failed to initiate task loop: failed to generate response"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			chatModel := new(MockChatModel)
			chatSession := new(MockChatSession)
			conversationService := new(MockConversationService)
			conversationService.markEyesWaitGroup = &sync.WaitGroup{}
			conversationService.markEyesWaitGroup.Add(1)
			fileQueryService := new(MockFileQueryService)
			sourceRepository := new(MockSourceRepository)
			ragCorpus := new(MockRAGCorpus)

			// モックの設定
			responseFormatter := new(MockResponseFormatter)
			tt.setupMocks(chatModel, chatSession, conversationService, fileQueryService, sourceRepository, ragCorpus, responseFormatter)

			// ConversationUsecaseの作成
			var usecase *ConversationUsecase
			if tt.disableRAG {
				usecase = NewConversationUsecase(
					chatModel,
					conversationService,
					fileQueryService,
					[]port.SourceRepository{sourceRepository},
					responseFormatter,
				)
			} else {
				usecase = NewConversationUsecase(
					chatModel,
					conversationService,
					fileQueryService,
					[]port.SourceRepository{sourceRepository},
					responseFormatter,
					WithConversationRAGCorpus(ragCorpus),
				)
			}

			// テスト実行
			err := usecase.Execute(context.Background())

			// 非同期処理の完了を待つ
			conversationService.markEyesWaitGroup.Wait()

			// 結果検証
			if tt.expectedError != nil {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedError.Error())
			} else {
				assert.NoError(t, err)
			}

			// モックの検証
			chatModel.AssertExpectations(t)
			chatSession.AssertExpectations(t)
			conversationService.AssertExpectations(t)
			fileQueryService.AssertExpectations(t)
			sourceRepository.AssertExpectations(t)
			ragCorpus.AssertExpectations(t)
		})
	}
}
