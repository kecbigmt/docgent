package application

import (
	"context"
	"docgent-backend/internal/domain/autoagent"
	"docgent-backend/internal/workflow"
	"fmt"
	"strconv"

	"github.com/google/go-github/v68/github"
	"go.uber.org/fx"
	"go.uber.org/zap"
)

type GitHubIssueCommentEventConsumerParams struct {
	fx.In

	AutoAgent       autoagent.Agent
	Logger          *zap.Logger
	ServiceProvider GitHubServiceProvider
}

type GitHubIssueCommentEventConsumer struct {
	agent           autoagent.Agent
	logger          *zap.Logger
	serviceProvider GitHubServiceProvider
}

func NewGitHubIssueCommentEventConsumer(params GitHubIssueCommentEventConsumerParams) *GitHubIssueCommentEventConsumer {
	return &GitHubIssueCommentEventConsumer{
		agent:           params.AutoAgent,
		logger:          params.Logger,
		serviceProvider: params.ServiceProvider,
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
	conversationService := c.serviceProvider.NewIssueCommentConversationService(installationID, ownerName, repoName, ev.Issue.GetNumber())

	// Get PR head branch using service provider
	headBranch, err := c.serviceProvider.GetPullRequestHeadBranch(ctx, installationID, ownerName, repoName, ev.Issue.GetNumber())
	if err != nil {
		c.logger.Error("Failed to get pull request head branch", zap.Error(err))
		return
	}

	// Create file query service with PR's head branch
	fileQueryService := c.serviceProvider.NewFileQueryService(installationID, ownerName, repoName, headBranch)

	// Create proposal service
	proposalService := c.serviceProvider.NewPullRequestAPI(installationID, ownerName, repoName, defaultBranch)

	// Create workflow instance
	workflow := workflow.NewProposalRefineWorkflow(
		c.agent,             // AI interaction
		conversationService, // Comment management
		fileQueryService,    // File operations
		proposalService,     // PR management
	)

	// Process feedback
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
