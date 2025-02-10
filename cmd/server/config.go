package main

import (
	"docgent-backend/internal/infrastructure/handler"
	"encoding/json"
	"os"
	"strconv"
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

func NewApplicationConfigServiceFromJSON() handler.ApplicationConfigService {
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

func NewApplicationConfigServiceFromEnv() handler.ApplicationConfigService {
	slackWorkspaceID := os.Getenv("SLACK_WORKSPACE_ID")
	if slackWorkspaceID == "" {
		panic("SLACK_WORKSPACE_ID is not set")
	}

	githubOwner := os.Getenv("GITHUB_OWNER")
	if githubOwner == "" {
		panic("GITHUB_OWNER is not set")
	}

	githubRepo := os.Getenv("GITHUB_REPO")
	if githubRepo == "" {
		panic("GITHUB_REPO is not set")
	}

	githubDefaultBranch := os.Getenv("GITHUB_DEFAULT_BRANCH")
	if githubDefaultBranch == "" {
		githubDefaultBranch = "main"
	}

	githubInstallationIDStr := os.Getenv("GITHUB_INSTALLATION_ID")
	if githubInstallationIDStr == "" {
		panic("GITHUB_INSTALLATION_ID is not set")
	}
	githubInstallationID, err := strconv.ParseInt(githubInstallationIDStr, 10, 64)
	if err != nil {
		panic("GITHUB_INSTALLATION_ID is not a valid integer")
	}

	vertexaiRagCorpusIDStr := os.Getenv("VERTEXAI_RAG_CORPUS_ID")
	var vertexaiRagCorpusID int64
	if vertexaiRagCorpusIDStr == "" {
		panic("VERTEXAI_RAG_CORPUS_ID is not set")
	} else {
		vertexaiRagCorpusID, err = strconv.ParseInt(vertexaiRagCorpusIDStr, 10, 64)
		if err != nil {
			panic("VERTEXAI_RAG_CORPUS_ID is not a valid integer")
		}
	}

	workspaces := []handler.Workspace{
		{
			SlackWorkspaceID:     slackWorkspaceID,
			GitHubOwner:          githubOwner,
			GitHubRepo:           githubRepo,
			GitHubInstallationID: githubInstallationID,
			VertexAICorpusID:     vertexaiRagCorpusID,
		},
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
