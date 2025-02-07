package github

import (
	"docgent-backend/internal/domain"
)

type PullRequestAPIFactory struct {
	api *API
}

func NewPullRequestAPIFactory(api *API) *PullRequestAPIFactory {
	return &PullRequestAPIFactory{api: api}
}

func (f *PullRequestAPIFactory) New(installationID int64, owner, repo, baseBranch, headBranch string) domain.ProposalRepository {
	client := f.api.NewClient(installationID)
	return NewPullRequestAPI(client, owner, repo, baseBranch, headBranch)
}
