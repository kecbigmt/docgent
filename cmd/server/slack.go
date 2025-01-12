package main

import (
	"os"

	"docgent-backend/internal/application"
	"docgent-backend/internal/infrastructure/slack"
)

func NewSlackAPI() application.SlackAPI {
	token := os.Getenv("SLACK_BOT_TOKEN")
	if token == "" {
		panic("SLACK_BOT_TOKEN is not set")
	}

	signingSecret := os.Getenv("SLACK_SIGNING_SECRET")
	if signingSecret == "" {
		panic("SLACK_SIGNING_SECRET is not set")
	}

	return slack.NewAPI(token, signingSecret)
}
