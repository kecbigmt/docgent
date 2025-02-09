package main

import (
	"docgent-backend/internal/infrastructure/handler"
	"encoding/json"
	"os"
)

type applicationConfigService struct {
	workspaces []handler.Workspace
}

func newApplicationConfigService(workspaces []handler.Workspace) handler.ApplicationConfigService {
	return &applicationConfigService{
		workspaces: workspaces,
	}
}

type config struct {
	Workspaces []handler.Workspace `json:"workspaces"`
}

func newApplicationConfigServiceFromEnv() handler.ApplicationConfigService {
	configBytes, err := os.ReadFile("config.json")
	if err != nil {
		panic(err)
	}

	var config config
	json.Unmarshal(configBytes, &config)

	workspaces := config.Workspaces

	for i, workspace := range workspaces {
		if workspace.GitHubDefaultBranch == "" {
			workspaces[i].GitHubDefaultBranch = "main"
		}
	}

	return newApplicationConfigService(workspaces)
}

func (s *applicationConfigService) GetWorkspaceBySlackWorkspaceID(slackWorkspaceID string) (handler.Workspace, error) {
	for _, workspace := range s.workspaces {
		if workspace.SlackWorkspaceID == slackWorkspaceID {
			return workspace, nil
		}
	}

	return handler.Workspace{}, handler.ErrWorkspaceNotFound
}

func (s *applicationConfigService) GetWorkspaceByGitHubInstallationID(githubInstallationID int64) (handler.Workspace, error) {
	for _, workspace := range s.workspaces {
		if workspace.GitHubInstallationID == githubInstallationID {
			return workspace, nil
		}
	}

	return handler.Workspace{}, handler.ErrWorkspaceNotFound
}
