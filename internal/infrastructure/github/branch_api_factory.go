package github

import (
	"docgent-backend/internal/application"
	"docgent-backend/internal/domain"
)

type BranchAPIFactory struct {
	api *API
}

func NewBranchAPIFactory(api *API) *BranchAPIFactory {
	return &BranchAPIFactory{
		api: api,
	}
}

func (f *BranchAPIFactory) New(params application.GitHubAppParams) domain.IncrementRepository {
	client := f.api.NewClient(params.InstallationID)
	return NewBranchAPI(client, params.Owner, params.Repo, params.DefaultBranch)
}
