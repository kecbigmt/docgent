package application

import (
	"context"

	"docgent-backend/internal/domain"
)

// GitHubServiceProvider defines the interface for creating GitHub-related services
type GitHubServiceProvider interface {
	// NewConversationService creates a new conversation service for the given context
	NewConversationService(ctx context.Context, params ConversationParams) (domain.ConversationService, error)

	// NewFileQueryService creates a new file query service for the given context
	NewFileQueryService(ctx context.Context, params FileQueryParams) (domain.FileQueryService, error)

	// NewPullRequestAPI creates a new pull request API for the given context
	NewPullRequestAPI(ctx context.Context, params PullRequestParams) (domain.ProposalService, error)

	// GetPullRequestHeadBranch gets the head branch of a pull request
	GetPullRequestHeadBranch(ctx context.Context, params PullRequestParams, number int) (string, error)
}

// ConversationParams contains conversation-specific context
type ConversationParams struct {
	InstallationID    int64
	Owner             string
	Repo              string
	PullRequestNumber int
	CommentID         int64
}

// FileQueryParams contains file query context
type FileQueryParams struct {
	InstallationID int64
	Owner          string
	Repo           string
	Branch         string
}

// PullRequestParams contains PR-specific context
type PullRequestParams struct {
	InstallationID int64
	Owner          string
	Repo           string
	DefaultBranch  string
}
