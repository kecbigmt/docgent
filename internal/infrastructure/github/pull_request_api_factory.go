package github

import (
	"docgent-backend/internal/application"
	"docgent-backend/internal/domain"
)

type PullRequestAPIFactory struct {
	api *API
}

func NewPullRequestAPIFactory(api *API) *PullRequestAPIFactory {
	return &PullRequestAPIFactory{api: api}
}

func (f *PullRequestAPIFactory) New(params application.GitHubAppParams) domain.ProposalRepository {
	client := f.api.NewClient(params.InstallationID)
	return NewPullRequestAPI(client, params.Owner, params.Repo)
}
