package application

import (
	"fmt"
	"strconv"

	"github.com/google/go-github/v68/github"
	"go.uber.org/fx"
	"go.uber.org/zap"
)

type GitHubIssueCommentEventConsumerParams struct {
	fx.In

	Logger                      *zap.Logger
	GitHubPullRequestAPIFactory GitHubPullRequestAPIFactory
}

type GitHubIssueCommentEventConsumer struct {
	logger                      *zap.Logger
	githubPullRequestAPIFactory GitHubPullRequestAPIFactory
}

func NewGitHubIssueCommentEventConsumer(params GitHubIssueCommentEventConsumerParams) *GitHubIssueCommentEventConsumer {
	return &GitHubIssueCommentEventConsumer{
		logger:                      params.Logger,
		githubPullRequestAPIFactory: params.GitHubPullRequestAPIFactory,
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

	githubPullRequestAPI := c.githubPullRequestAPIFactory.New(GitHubAppParams{
		InstallationID: installationID,
		Repo:           repoName,
		Owner:          ownerName,
		DefaultBranch:  defaultBranch,
	})

	handle := githubPullRequestAPI.NewProposalHandle(strconv.Itoa(ev.Issue.GetNumber()))

	comment, err := githubPullRequestAPI.CreateComment(handle, ev.Comment.GetBody())
	if err != nil {
		c.logger.Error("Failed to create comment", zap.Error(err), zap.String("pull_request", pullRequestPath))
		return
	}

	c.logger.Info(
		"Comment created",
		zap.String("comment", pullRequestPath+"#issuecomment-"+comment.Handle.Value),
	)
}
