package github

import (
	"docgent-backend/internal/application"
	"docgent-backend/internal/domain"
)

type DocumentRepositoryFactory struct {
	api *API
}

func NewDocumentRepositoryFactory(api *API) *DocumentRepositoryFactory {
	return &DocumentRepositoryFactory{
		api: api,
	}
}

func (f *DocumentRepositoryFactory) New(params application.GitHubAppParams) domain.DocumentRepository {
	client := f.api.NewClient(params.InstallationID)
	return NewDocumentRepository(client, params.Owner, params.Repo, params.BaseBranch)
}
