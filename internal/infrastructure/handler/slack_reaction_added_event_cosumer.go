package handler

import (
	"context"
	"fmt"
	"time"

	"github.com/slack-go/slack/slackevents"
	"go.uber.org/fx"
	"go.uber.org/zap"

	"docgent/internal/application"
	"docgent/internal/application/port"
	"docgent/internal/domain"
	"docgent/internal/infrastructure/github"
	"docgent/internal/infrastructure/slack"
)

type SlackReactionAddedEventConsumerParams struct {
	fx.In

	Logger                   *zap.Logger
	SlackAPI                 *slack.API
	GitHubServiceProvider    *github.ServiceProvider
	SlackServiceProvider     *slack.ServiceProvider
	ChatModel                domain.ChatModel
	RAGService               port.RAGService
	ApplicationConfigService ApplicationConfigService
}

type SlackReactionAddedEventConsumer struct {
	logger                   *zap.Logger
	slackAPI                 *slack.API
	githubServiceProvider    *github.ServiceProvider
	slackServiceProvider     *slack.ServiceProvider
	chatModel                domain.ChatModel
	ragService               port.RAGService
	applicationConfigService ApplicationConfigService
}

func NewSlackReactionAddedEventConsumer(params SlackReactionAddedEventConsumerParams) *SlackReactionAddedEventConsumer {
	return &SlackReactionAddedEventConsumer{
		logger:                   params.Logger,
		slackAPI:                 params.SlackAPI,
		githubServiceProvider:    params.GitHubServiceProvider,
		slackServiceProvider:     params.SlackServiceProvider,
		chatModel:                params.ChatModel,
		ragService:               params.RAGService,
		applicationConfigService: params.ApplicationConfigService,
	}
}

func (h *SlackReactionAddedEventConsumer) EventType() string {
	return "reaction_added"
}

func (h *SlackReactionAddedEventConsumer) ConsumeEvent(event slackevents.EventsAPIInnerEvent, workspace Workspace) {
	ev, ok := event.Data.(*slackevents.ReactionAddedEvent)
	if !ok {
		h.logger.Error("Failed to convert event data to ReactionAddedEvent")
		return
	}

	if ev.Reaction != "doc_it" {
		h.logger.Info("Reaction is not doc_it", zap.String("reaction", ev.Reaction))
		return
	}

	threadTimestamp := ev.Item.Timestamp

	ref := slack.NewConversationRef(workspace.SlackWorkspaceID, ev.Item.Channel, threadTimestamp, threadTimestamp)
	conversationService := h.slackServiceProvider.NewConversationService(ref, ev.User)

	ctx := context.Background()
	baseBranchName := workspace.GitHubDefaultBranch
	newBranchName := fmt.Sprintf("docgent/%d", time.Now().Unix())

	branchService := h.githubServiceProvider.NewBranchService(workspace.GitHubInstallationID, workspace.GitHubOwner, workspace.GitHubRepo)
	err := branchService.CreateBranch(ctx, baseBranchName, newBranchName)
	if err != nil {
		h.logger.Error("Failed to create branch", zap.Error(err))
		conversationService.Reply(":warning: エラー: ブランチの作成に失敗しました", true)
		return
	}

	fileQueryService := h.githubServiceProvider.NewFileQueryService(workspace.GitHubInstallationID, workspace.GitHubOwner, workspace.GitHubRepo, newBranchName)
	fileRepository := h.githubServiceProvider.NewFileRepository(workspace.GitHubInstallationID, workspace.GitHubOwner, workspace.GitHubRepo, newBranchName)

	sourceRepositories := []port.SourceRepository{
		h.slackServiceProvider.NewSourceRepository(),
		h.githubServiceProvider.NewSourceRepository(workspace.GitHubInstallationID),
	}

	githubPullRequestAPI := h.githubServiceProvider.NewPullRequestAPI(workspace.GitHubInstallationID, workspace.GitHubOwner, workspace.GitHubRepo, baseBranchName, newBranchName)

	var options []application.NewProposalGenerateUsecaseOption
	// If VertexAICorpusID is set, use RAG corpus
	if workspace.VertexAICorpusID > 0 {
		options = append(options, application.WithProposalGenerateRAGCorpus(h.ragService.GetCorpus(workspace.VertexAICorpusID)))
	}

	// Slackのレスポンスフォーマッターを取得
	responseFormatter := h.slackServiceProvider.NewResponseFormatter()

	// ドキュメントを生成
	proposalGenerateUsecase := application.NewProposalGenerateUsecase(
		h.chatModel,
		conversationService,
		fileQueryService,
		fileRepository,
		sourceRepositories,
		githubPullRequestAPI,
		responseFormatter,
		options...,
	)
	proposalHandle, err := proposalGenerateUsecase.Execute(ctx)
	if err != nil {
		h.logger.Error("Failed to generate increment", zap.Error(err))
		conversationService.Reply(":warning: エラー: ドキュメントの生成に失敗しました", true)
		return
	}

	// 成功メッセージを投稿
	conversationService.Reply(fmt.Sprintf(
		"PR: https://github.com/%s/%s/pull/%s",
		workspace.GitHubOwner,
		workspace.GitHubRepo,
		proposalHandle.Value,
	), false)
}
