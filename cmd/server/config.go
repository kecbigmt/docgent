package main

import (
	"docgent-backend/internal/application"
	"encoding/json"
	"os"
)

type ApplicationConfigService struct {
	workspaces []application.Workspace
}

func NewApplicationConfigService(workspaces []application.Workspace) application.ApplicationConfigService {
	return &ApplicationConfigService{
		workspaces: workspaces,
	}
}

type Config struct {
	Workspaces []application.Workspace `json:"workspaces"`
}

func NewApplicationConfigServiceFromEnv() application.ApplicationConfigService {
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

func (s *ApplicationConfigService) GetWorkspaceBySlackWorkspaceID(slackWorkspaceID string) (application.Workspace, error) {
	for _, workspace := range s.workspaces {
		if workspace.SlackWorkspaceID == slackWorkspaceID {
			return workspace, nil
		}
	}

	return application.Workspace{}, application.ErrWorkspaceNotFound
}

func (s *ApplicationConfigService) GetWorkspaceByGitHubInstallationID(githubInstallationID int64) (application.Workspace, error) {
	for _, workspace := range s.workspaces {
		if workspace.GitHubInstallationID == githubInstallationID {
			return workspace, nil
		}
	}

	return application.Workspace{}, application.ErrWorkspaceNotFound
}
