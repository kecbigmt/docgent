package application

import (
	"context"
	"fmt"
	"strconv"

	"github.com/google/go-github/v68/github"
	"go.uber.org/fx"
	"go.uber.org/zap"

	appgithub "docgent-backend/internal/application/github"
	"docgent-backend/internal/domain"
	"docgent-backend/internal/workflow"
)

type GitHubIssueCommentEventConsumerParams struct {
	fx.In

	ChatModel                domain.ChatModel
	Logger                   *zap.Logger
	GitHubServiceProvider    appgithub.ServiceProvider
	RAGService               domain.RAGService
	ApplicationConfigService ApplicationConfigService
}

type GitHubIssueCommentEventConsumer struct {
	chatModel                domain.ChatModel
	logger                   *zap.Logger
	githubServiceProvider    appgithub.ServiceProvider
	ragService               domain.RAGService
	applicationConfigService ApplicationConfigService
}

func NewGitHubIssueCommentEventConsumer(params GitHubIssueCommentEventConsumerParams) *GitHubIssueCommentEventConsumer {
	return &GitHubIssueCommentEventConsumer{
		chatModel:                params.ChatModel,
		logger:                   params.Logger,
		githubServiceProvider:    params.GitHubServiceProvider,
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
	conversationService := c.githubServiceProvider.NewIssueCommentConversationService(installationID, ownerName, repoName, ev.Issue.GetNumber())

	// Get PR head branch using service provider
	headBranch, err := c.githubServiceProvider.GetPullRequestHeadBranch(ctx, installationID, ownerName, repoName, ev.Issue.GetNumber())
	if err != nil {
		c.logger.Error("Failed to get pull request head branch", zap.Error(err))
		return
	}

	// Create file query service with PR's head branch
	fileQueryService := c.githubServiceProvider.NewFileQueryService(installationID, ownerName, repoName, headBranch)

	// Create file change service
	fileChangeService := c.githubServiceProvider.NewFileChangeService(installationID, ownerName, repoName, headBranch)

	// Create proposal service
	// TODO: PRの作成以外ではブランチ名が不要なので、サービスを分ける
	proposalService := c.githubServiceProvider.NewPullRequestAPI(installationID, ownerName, repoName, defaultBranch, "")

	// Create workflow instance
	workflow := workflow.NewProposalRefineWorkflow(
		c.chatModel,         // AI interaction
		conversationService, // Comment management
		fileQueryService,    // File operations
		fileChangeService,   // File operations
		proposalService,     // PR management
		c.ragService.GetCorpus(workspace.VertexAICorpusID), // RAG corpus
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
