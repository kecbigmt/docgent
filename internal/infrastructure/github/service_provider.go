package github

import (
	"context"
	"fmt"

	"docgent/internal/application/port"
	"docgent/internal/domain"
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
func (p *ServiceProvider) NewIssueCommentConversationService(installationID int64, ref *IssueCommentRef, fromUserID string) port.ConversationService {
	return NewIssueCommentConversationService(
		p.api.NewClient(installationID),
		ref,
		fromUserID,
	)
}

// NewReviewCommentConversationService creates a conversation service with the proper context
func (p *ServiceProvider) NewReviewCommentConversationService(installationID int64, owner, repo string, prNumber int, sourceCommentID int64, fromUserID string) port.ConversationService {
	return NewReviewCommentConversationService(
		p.api.NewClient(installationID),
		owner,
		repo,
		prNumber,
		sourceCommentID,
		fromUserID,
	)
}

// NewFileQueryService creates a file query service with the proper context
func (p *ServiceProvider) NewFileQueryService(installationID int64, owner, repo, branch string) port.FileQueryService {
	return NewFileQueryService(p.api.NewClient(installationID), owner, repo, branch)
}

// NewFileRepository creates a file repository with the proper context
func (p *ServiceProvider) NewFileRepository(installationID int64, owner, repo, branch string) *FileRepository {
	return NewFileRepository(p.api.NewClient(installationID), owner, repo, branch)
}

// NewSourceRepository creates a source repository with the proper context
func (p *ServiceProvider) NewSourceRepository(installationID int64) *SourceRepository {
	return NewSourceRepository(p.api.NewClient(installationID))
}

// NewBranchService creates a branch service with the proper context
func (p *ServiceProvider) NewBranchService(installationID int64, owner, repo string) *BranchService {
	return NewBranchService(p.api.NewClient(installationID), owner, repo)
}

// NewPullRequestAPI creates a pull request API with the proper context
func (p *ServiceProvider) NewPullRequestAPI(installationID int64, owner, repo, baseBranch, headBranch string) domain.ProposalRepository {
	return NewPullRequestAPI(p.api.NewClient(installationID), owner, repo, baseBranch, headBranch)
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

// NewResponseFormatter creates a response formatter for GitHub
func (p *ServiceProvider) NewResponseFormatter() port.ResponseFormatter {
	return NewResponseFormatter()
}
