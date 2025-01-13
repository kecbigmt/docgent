package application

import "docgent-backend/internal/domain"

type GitHubPullRequestAPIFactory interface {
	New(params GitHubAppParams) domain.ProposalRepository
}
