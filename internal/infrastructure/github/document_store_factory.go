package github

import (
	"docgent-backend/internal/application"
	"docgent-backend/internal/domain"
)

type DocumentStoreFactory struct {
	api *API
}

func NewDocumentStoreFactory(api *API) *DocumentStoreFactory {
	return &DocumentStoreFactory{
		api: api,
	}
}

func (f *DocumentStoreFactory) New(params application.GitHubAppParams) domain.DocumentStore {
	client := f.api.NewClient(params.InstallationID)
	return NewDocumentStore(client, params.Owner, params.Repo, params.BaseBranch)
}
