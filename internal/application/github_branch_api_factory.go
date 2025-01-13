package application

import "docgent-backend/internal/domain"

type GitHubBranchAPIFactory interface {
	New(params GitHubAppParams) domain.IncrementRepository
}
