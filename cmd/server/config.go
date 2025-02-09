package main

import (
	"docgent-backend/internal/infrastructure/handler"
	"encoding/json"
	"os"
)

type ApplicationConfigService struct {
	workspaces []handler.Workspace
}

func NewApplicationConfigService(workspaces []handler.Workspace) handler.ApplicationConfigService {
	return &ApplicationConfigService{
		workspaces: workspaces,
	}
}

type Config struct {
	Workspaces []handler.Workspace `json:"workspaces"`
}

func NewApplicationConfigServiceFromEnv() handler.ApplicationConfigService {
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

func (s *ApplicationConfigService) GetWorkspaceBySlackWorkspaceID(slackWorkspaceID string) (handler.Workspace, error) {
	for _, workspace := range s.workspaces {
		if workspace.SlackWorkspaceID == slackWorkspaceID {
			return workspace, nil
		}
	}

	return handler.Workspace{}, handler.ErrWorkspaceNotFound
}

func (s *ApplicationConfigService) GetWorkspaceByGitHubInstallationID(githubInstallationID int64) (handler.Workspace, error) {
	for _, workspace := range s.workspaces {
		if workspace.GitHubInstallationID == githubInstallationID {
			return workspace, nil
		}
	}

	return handler.Workspace{}, handler.ErrWorkspaceNotFound
}
