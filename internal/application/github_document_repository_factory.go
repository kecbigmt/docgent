package application

import "docgent-backend/internal/domain"

type GitHubDocumentRepositoryFactory interface {
	New(params GitHubAppParams) domain.DocumentRepository
}
