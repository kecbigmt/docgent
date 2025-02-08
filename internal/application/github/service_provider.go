package github

import (
	"context"
	"docgent-backend/internal/domain"
)

// ServiceProvider defines the interface for creating GitHub-related services
type ServiceProvider interface {
	NewIssueCommentConversationService(installationID int64, owner, repo string, prNumber int) domain.ConversationService

	NewReviewCommentConversationService(installationID int64, owner, repo string, prNumber int, parentCommentID int64) domain.ConversationService

	NewFileQueryService(installationID int64, owner, repo, branch string) domain.FileQueryService

	NewFileChangeService(installationID int64, owner, repo, branch string) domain.FileChangeService

	NewBranchService(installationID int64, owner, repo string) BranchService

	NewPullRequestAPI(installationID int64, owner, repo, baseBranch, headBranch string) domain.ProposalRepository

	GetPullRequestHeadBranch(ctx context.Context, installationID int64, owner, repo string, number int) (string, error)
}
