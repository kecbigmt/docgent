package handler

import (
	"context"
	"fmt"
	"strconv"

	"github.com/google/go-github/v68/github"
	"go.uber.org/fx"
	"go.uber.org/zap"

	"docgent/internal/application"
	"docgent/internal/application/port"
	"docgent/internal/domain"
	infragithub "docgent/internal/infrastructure/github"
	"docgent/internal/infrastructure/slack"
)

type GitHubIssueCommentEventConsumerParams struct {
	fx.In

	ChatModel                domain.ChatModel
	Logger                   *zap.Logger
	GitHubServiceProvider    *infragithub.ServiceProvider
	SlackServiceProvider     *slack.ServiceProvider
	RAGService               port.RAGService
	ApplicationConfigService ApplicationConfigService
}

type GitHubIssueCommentEventConsumer struct {
	chatModel                domain.ChatModel
	logger                   *zap.Logger
	githubServiceProvider    *infragithub.ServiceProvider
	slackServiceProvider     *slack.ServiceProvider
	ragService               port.RAGService
	applicationConfigService ApplicationConfigService
}

func NewGitHubIssueCommentEventConsumer(params GitHubIssueCommentEventConsumerParams) *GitHubIssueCommentEventConsumer {
	return &GitHubIssueCommentEventConsumer{
		chatModel:                params.ChatModel,
		logger:                   params.Logger,
		githubServiceProvider:    params.GitHubServiceProvider,
		slackServiceProvider:     params.SlackServiceProvider,
		ragService:               params.RAGService,
		applicationConfigService: params.ApplicationConfigService,
	}
}

func (c *GitHubIssueCommentEventConsumer) EventType() string {
	return "issue_comment"
}

func (c *GitHubIssueCommentEventConsumer) ConsumeEvent(event interface{}) {
	ev, ok := event.(*github.IssueCommentEvent)
	if !ok {
		c.logger.Error("Failed to convert event data to IssueCommentEvent")
		return
	}
	c.logger.Info("Processing issue comment event", zap.String("action", ev.GetAction()))

	installationID := ev.GetInstallation().GetID()
	repo := ev.GetRepo()
	repoName := repo.GetName()
	defaultBranch := repo.GetDefaultBranch()
	ownerName := repo.GetOwner().GetLogin()
	issueNumber := strconv.Itoa(ev.Issue.GetNumber())

	workspace, err := c.applicationConfigService.GetWorkspaceByGitHubInstallationID(installationID)
	if err != nil {
		if err == ErrWorkspaceNotFound {
			c.logger.Warn("Unknown GitHub installation ID", zap.Int64("installation_id", installationID))
			return
		}
		c.logger.Error("Failed to get workspace", zap.Error(err))
		return
	}

	// Skip if this is not a pull request
	if ev.Issue.PullRequestLinks == nil {
		c.logger.Debug(
			"Skipping non-pull request comment",
			zap.String("issue", fmt.Sprintf("https://github.com/%s/%s/issues/%s", ownerName, repoName, issueNumber)),
		)
		return
	}

	pullRequestPath := fmt.Sprintf("https://github.com/%s/%s/pull/%s", ownerName, repoName, issueNumber)

	if ev.Sender.GetType() != "User" {
		c.logger.Debug(
			"Skipping non-user comment",
			zap.String("pull_request", pullRequestPath),
		)
		return
	}

	ctx := context.Background()

	// Create conversation service with PR and comment context
	commentID := ev.Comment.GetID()
	ref := infragithub.NewIssueCommentRef(ownerName, repoName, ev.Issue.GetNumber(), commentID)
	conversationService := c.githubServiceProvider.NewIssueCommentConversationService(installationID, ref)

	// Get PR head branch using service provider
	headBranch, err := c.githubServiceProvider.GetPullRequestHeadBranch(ctx, installationID, ownerName, repoName, ev.Issue.GetNumber())
	if err != nil {
		c.logger.Error("Failed to get pull request head branch", zap.Error(err))
		return
	}

	// Create file query service with PR's head branch
	fileQueryService := c.githubServiceProvider.NewFileQueryService(installationID, ownerName, repoName, headBranch)

	// Create file change service
	fileRepository := c.githubServiceProvider.NewFileRepository(installationID, ownerName, repoName, headBranch)

	sourceRepositories := []port.SourceRepository{
		c.githubServiceProvider.NewSourceRepository(installationID),
		c.slackServiceProvider.NewSourceRepository(),
	}

	// Create proposal service
	// TODO: PRの作成以外ではブランチ名が不要なので、サービスを分ける
	proposalService := c.githubServiceProvider.NewPullRequestAPI(installationID, ownerName, repoName, defaultBranch, "")

	var options []application.NewProposalRefineUsecaseOption
	// If VertexAICorpusID is set, use RAG corpus
	if workspace.VertexAICorpusID > 0 {
		options = append(options, application.WithProposalRefineRAGCorpus(c.ragService.GetCorpus(workspace.VertexAICorpusID)))
	}

	// Create workflow instance
	workflow := application.NewProposalRefineUsecase(
		c.chatModel,         // AI interaction
		conversationService, // Comment management
		fileQueryService,    // File operations
		fileRepository,      // File operations
		sourceRepositories,
		proposalService, // PR management
		options...,
	)

	// Process feedbacks
	handle := proposalService.NewProposalHandle(strconv.Itoa(ev.Issue.GetNumber()))
	if err := workflow.Refine(handle, ev.Comment.GetBody()); err != nil {
		c.logger.Error("Refinement failed", zap.Error(err))
		return
	}

	c.logger.Info(
		"Comment processed and refinement applied",
		zap.String("pull_request", pullRequestPath),
	)
}
