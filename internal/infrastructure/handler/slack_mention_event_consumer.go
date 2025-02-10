package handler

import (
	"docgent-backend/internal/application"
	"docgent-backend/internal/application/port"
	"docgent-backend/internal/domain"
	"docgent-backend/internal/infrastructure/slack"

	"github.com/slack-go/slack/slackevents"
	"go.uber.org/fx"
	"go.uber.org/zap"
)

type SlackMentionEventConsumerParams struct {
	fx.In

	Logger               *zap.Logger
	ChatModel            domain.ChatModel
	RAGService           port.RAGService
	SlackServiceProvider *slack.ServiceProvider
}

type SlackMentionEventConsumer struct {
	log                  *zap.Logger
	chatModel            domain.ChatModel
	ragService           port.RAGService
	slackServiceProvider *slack.ServiceProvider
}

func NewSlackMentionEventConsumer(params SlackMentionEventConsumerParams) *SlackMentionEventConsumer {
	return &SlackMentionEventConsumer{
		log:                  params.Logger,
		chatModel:            params.ChatModel,
		ragService:           params.RAGService,
		slackServiceProvider: params.SlackServiceProvider,
	}
}

func (c *SlackMentionEventConsumer) EventType() string {
	return string(slackevents.AppMention)
}

func (c *SlackMentionEventConsumer) ConsumeEvent(event slackevents.EventsAPIInnerEvent, workspace Workspace) {
	appMentionEvent, ok := event.Data.(*slackevents.AppMentionEvent)
	if !ok {
		c.log.Error("Failed to convert event to AppMentionEvent")
		return
	}

	question := appMentionEvent.Text
	threadTimestamp := appMentionEvent.ThreadTimeStamp
	sourceMessageTimestamp := appMentionEvent.TimeStamp
	if threadTimestamp == "" {
		threadTimestamp = sourceMessageTimestamp
	}

	// 会話サービスを初期化
	conversationService := c.slackServiceProvider.NewConversationService(appMentionEvent.Channel, threadTimestamp, sourceMessageTimestamp)

	var options []application.NewQuestionAnswerUsecaseOption
	// If VertexAICorpusID is set, use RAG corpus
	if workspace.VertexAICorpusID > 0 {
		options = append(options, application.WithQuestionAnswerRAGCorpus(c.ragService.GetCorpus(workspace.VertexAICorpusID)))
	}

	// QuestionAnswerUsecaseを初期化
	questionAnswerUsecase := application.NewQuestionAnswerUsecase(
		c.chatModel,
		conversationService,
		options...,
	)

	err := questionAnswerUsecase.Execute(question)
	if err != nil {
		c.log.Error("Failed to execute question answer usecase", zap.Error(err))
		conversationService.Reply(":warning: エラー: 質問への回答に失敗しました")
		return
	}
}
