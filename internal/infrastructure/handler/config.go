package handler

import "errors"

type Workspace struct {
	SlackWorkspaceID     string `json:"slack_workspace_id"`
	GitHubInstallationID int64  `json:"github_installation_id"`
	GitHubOwner          string `json:"github_owner"`
	GitHubRepo           string `json:"github_repo"`
	GitHubDefaultBranch  string `json:"github_default_branch"`
	VertexAICorpusID     string `json:"vertexai_rag_corpus_id"`
}

var ErrWorkspaceNotFound = errors.New("workspace not found")

type ApplicationConfigService interface {
	GetWorkspaceBySlackWorkspaceID(slackWorkspaceID string) (Workspace, error)
	GetWorkspaceByGitHubInstallationID(githubInstallationID int64) (Workspace, error)
}
