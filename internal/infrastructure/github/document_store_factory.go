package github

import (
	"docgent-backend/internal/application"
	"docgent-backend/internal/model/infrastructure"
)

type DocumentStoreFactory struct {
	api *API
}

func NewDocumentStoreFactory(api *API) *DocumentStoreFactory {
	return &DocumentStoreFactory{
		api: api,
	}
}

func (f *DocumentStoreFactory) New(params application.GitHubAppParams) infrastructure.DocumentStore {
	client := f.api.NewClient(params.InstallationID)
	return NewDocumentStore(client, params.Owner, params.Repo, params.BaseBranch)
}
