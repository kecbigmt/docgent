package application

import (
	"context"
	"docgent-backend/internal/domain"
)

// GitHubServiceProvider defines the interface for creating GitHub-related services
type GitHubServiceProvider interface {
	NewIssueCommentConversationService(installationID int64, owner, repo string, prNumber int) domain.ConversationService

	NewReviewCommentConversationService(installationID int64, owner, repo string, prNumber int, parentCommentID int64) domain.ConversationService

	NewFileQueryService(installationID int64, owner, repo, branch string) domain.FileQueryService

	NewPullRequestAPI(installationID int64, owner, repo, baseBranch string) domain.ProposalRepository

	GetPullRequestHeadBranch(ctx context.Context, installationID int64, owner, repo string, number int) (string, error)
}
