package github

import (
	"context"
	"fmt"

	"docgent-backend/internal/domain"
	"docgent-backend/internal/domain/autoagent"
)

// ServiceProvider implements the GitHubServiceProvider interface
type ServiceProvider struct {
	api *API
}

func NewServiceProvider(api *API) *ServiceProvider {
	return &ServiceProvider{
		api: api,
	}
}

// NewIssueCommentConversationService creates a conversation service with the proper context
func (p *ServiceProvider) NewIssueCommentConversationService(installationID int64, owner, repo string, prNumber int) autoagent.ConversationService {
	return NewIssueCommentConversationService(
		p.api.NewClient(installationID),
		owner,
		repo,
		prNumber,
	)
}

// NewReviewCommentConversationService creates a conversation service with the proper context
func (p *ServiceProvider) NewReviewCommentConversationService(installationID int64, owner, repo string, prNumber int, parentCommentID int64) autoagent.ConversationService {
	return NewReviewCommentConversationService(
		p.api.NewClient(installationID),
		owner,
		repo,
		prNumber,
		parentCommentID,
	)
}

// NewFileQueryService creates a file query service with the proper context
func (p *ServiceProvider) NewFileQueryService(installationID int64, owner, repo, branch string) domain.FileQueryService {
	return NewFileQueryService(p.api.NewClient(installationID), owner, repo, branch)
}

// NewPullRequestAPI creates a pull request API with the proper context
func (p *ServiceProvider) NewPullRequestAPI(installationID int64, owner, repo, baseBranch string) domain.ProposalRepository {
	return NewPullRequestAPI(p.api.NewClient(installationID), owner, repo, baseBranch)
}

// GetPullRequestHeadBranch gets the head branch of a pull request
func (p *ServiceProvider) GetPullRequestHeadBranch(ctx context.Context, installationID int64, owner, repo string, number int) (string, error) {
	client := p.api.NewClient(installationID)

	pr, _, err := client.PullRequests.Get(ctx, owner, repo, number)
	if err != nil {
		return "", fmt.Errorf("failed to get pull request details: %w", err)
	}
	return pr.Head.GetRef(), nil
}
