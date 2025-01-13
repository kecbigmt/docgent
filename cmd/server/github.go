package main

import (
	"log"
	"os"
	"strconv"

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

func NewGitHubAPI() github.API {
	appIDStr := os.Getenv("GITHUB_APP_ID")
	if appIDStr == "" {
		log.Fatal("GITHUB_APP_PRIVATE_KEY is not set")
	}
	appID, err := strconv.ParseInt(appIDStr, 10, 64)
	if err != nil {
		log.Fatalf("GITHUB_APP_ID is invalid: %v", err)
	}

	privateKey := os.Getenv("GITHUB_APP_PRIVATE_KEY")
	if privateKey == "" {
		log.Fatal("GITHUB_APP_PRIVATE_KEY is not set")
	}

	return *github.NewAPI(appID, []byte(privateKey))
}
