package application

import "docgent-backend/internal/domain"

type GitHubDocumentStoreFactory interface {
	New(params GitHubAppParams) domain.DocumentStore
}
