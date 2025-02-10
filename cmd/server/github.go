package main

import (
	"log"
	"os"
	"strconv"

	"docgent/internal/infrastructure/github"
)

func newGitHubAPI() *github.API {
	appIDStr := os.Getenv("GITHUB_APP_ID")
	if appIDStr == "" {
		log.Fatal("GITHUB_APP_ID is not set")
	}
	appID, err := strconv.ParseInt(appIDStr, 10, 64)
	if err != nil {
		log.Fatalf("GITHUB_APP_ID is invalid: %v", err)
	}

	privateKey := os.Getenv("GITHUB_APP_PRIVATE_KEY")
	if privateKey == "" {
		log.Fatal("GITHUB_APP_PRIVATE_KEY is not set")
	}

	return github.NewAPI(appID, []byte(privateKey))
}

func newGitHubWebhookRequestParser() *github.WebhookRequestParser {
	secret := os.Getenv("GITHUB_WEBHOOK_SECRET")
	if secret == "" {
		log.Fatal("GITHUB_WEBHOOK_SECRET is not set")
	}

	return github.NewWebhookRequestParser(secret)
}
