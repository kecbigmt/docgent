package handler

import (
	"context"
	"docgent/internal/application"
	"docgent/internal/application/port"
	"docgent/internal/domain"
	"docgent/internal/infrastructure/github"
	"docgent/internal/infrastructure/slack"

	"github.com/slack-go/slack/slackevents"
	"go.uber.org/fx"
	"go.uber.org/zap"
)

type SlackMentionEventConsumerParams struct {
	fx.In

	Logger                *zap.Logger
	ChatModel             domain.ChatModel
	RAGService            port.RAGService
	SlackServiceProvider  *slack.ServiceProvider
	GitHubServiceProvider *github.ServiceProvider
}

type SlackMentionEventConsumer struct {
	log                   *zap.Logger
	chatModel             domain.ChatModel
	ragService            port.RAGService
	slackServiceProvider  *slack.ServiceProvider
	githubServiceProvider *github.ServiceProvider
}

func NewSlackMentionEventConsumer(params SlackMentionEventConsumerParams) *SlackMentionEventConsumer {
	return &SlackMentionEventConsumer{
		log:                   params.Logger,
		chatModel:             params.ChatModel,
		ragService:            params.RAGService,
		slackServiceProvider:  params.SlackServiceProvider,
		githubServiceProvider: params.GitHubServiceProvider,
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

	threadTimestamp := appMentionEvent.ThreadTimeStamp
	sourceMessageTimestamp := appMentionEvent.TimeStamp
	if threadTimestamp == "" {
		threadTimestamp = sourceMessageTimestamp
	}

	// 会話サービスを初期化
	ref := slack.NewConversationRef(workspace.SlackWorkspaceID, appMentionEvent.Channel, threadTimestamp, sourceMessageTimestamp)
	conversationService := c.slackServiceProvider.NewConversationService(ref, appMentionEvent.User)

	ctx := context.Background()

	var options []application.NewConversationUsecaseOption
	// If VertexAICorpusID is set, use RAG corpus
	if workspace.VertexAICorpusID > 0 {
		options = append(options, application.WithConversationRAGCorpus(c.ragService.GetCorpus(workspace.VertexAICorpusID)))
	}

	sourceRepositories := []port.SourceRepository{
		c.slackServiceProvider.NewSourceRepository(),
		c.githubServiceProvider.NewSourceRepository(workspace.GitHubInstallationID),
	}
	fileQueryService := c.githubServiceProvider.NewFileQueryService(workspace.GitHubInstallationID, workspace.GitHubOwner, workspace.GitHubRepo, workspace.GitHubDefaultBranch)

	// ConversationUsecaseを初期化
	conversationUsecase := application.NewConversationUsecase(
		c.chatModel,
		conversationService,
		fileQueryService,
		sourceRepositories,
		options...,
	)

	err := conversationUsecase.Execute(ctx)
	if err != nil {
		c.log.Error("Failed to execute conversation usecase", zap.Error(err))
		conversationService.Reply(":warning: エラー: 会話の処理に失敗しました", false)
		return
	}
}
