package application

import "docgent-backend/internal/model/infrastructure"

type GitHubDocumentStoreFactory interface {
	New(params GitHubAppParams) infrastructure.DocumentStore
}
