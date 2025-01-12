package main

import (
	"os"

	"docgent-backend/internal/infrastructure/github"
)

func NewGitHubAPIConfig() github.APIConfig {
	token := os.Getenv("GITHUB_TOKEN")
	if token == "" {
		panic("GITHUB_TOKEN is not set")
	}

	owner := os.Getenv("GITHUB_OWNER")
	if owner == "" {
		panic("GITHUB_OWNER is not set")
	}

	repo := os.Getenv("GITHUB_REPO")
	if repo == "" {
		panic("GITHUB_REPO is not set")
	}

	baseBranch := os.Getenv("GITHUB_BASE_BRANCH")
	if baseBranch == "" {
		baseBranch = "main" // デフォルト値
	}

	return github.APIConfig{
		Token:      token,
		Owner:      owner,
		Repo:       repo,
		BaseBranch: baseBranch,
	}
}
