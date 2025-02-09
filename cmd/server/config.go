package main

import (
	"encoding/json"
	"errors"
	"os"
)

type Workspace struct {
	SlackWorkspaceID     string `json:"slack_workspace_id"`
	GitHubInstallationID int64  `json:"github_installation_id"`
	GitHubOwner          string `json:"github_owner"`
	GitHubRepo           string `json:"github_repo"`
	GitHubDefaultBranch  string `json:"github_default_branch"`
	VertexAICorpusID     string `json:"vertexai_rag_corpus_id"`
}

var ErrWorkspaceNotFound = errors.New("workspace not found")

type ApplicationConfigService struct {
	workspaces []Workspace
}

func NewApplicationConfigService(workspaces []Workspace) *ApplicationConfigService {
	return &ApplicationConfigService{
		workspaces: workspaces,
	}
}

type Config struct {
	Workspaces []Workspace `json:"workspaces"`
}

func NewApplicationConfigServiceFromEnv() *ApplicationConfigService {
	configBytes, err := os.ReadFile("config.json")
	if err != nil {
		panic(err)
	}

	var config Config
	json.Unmarshal(configBytes, &config)

	workspaces := config.Workspaces

	for i, workspace := range workspaces {
		if workspace.GitHubDefaultBranch == "" {
			workspaces[i].GitHubDefaultBranch = "main"
		}
	}

	return NewApplicationConfigService(workspaces)
}

func (s *ApplicationConfigService) GetWorkspaceBySlackWorkspaceID(slackWorkspaceID string) (Workspace, error) {
	for _, workspace := range s.workspaces {
		if workspace.SlackWorkspaceID == slackWorkspaceID {
			return workspace, nil
		}
	}

	return Workspace{}, ErrWorkspaceNotFound
}

func (s *ApplicationConfigService) GetWorkspaceByGitHubInstallationID(githubInstallationID int64) (Workspace, error) {
	for _, workspace := range s.workspaces {
		if workspace.GitHubInstallationID == githubInstallationID {
			return workspace, nil
		}
	}

	return Workspace{}, ErrWorkspaceNotFound
}
