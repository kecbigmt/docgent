package main

import (
	"os"

	"docgent-backend/internal/application"
)

func NewSlackAPIConfig() application.SlackAPIConfig {
	token := os.Getenv("SLACK_BOT_TOKEN")
	if token == "" {
		panic("SLACK_BOT_TOKEN is not set")
	}

	signingSecret := os.Getenv("SLACK_SIGNING_SECRET")
	if signingSecret == "" {
		panic("SLACK_SIGNING_SECRET is not set")
	}

	return application.SlackAPIConfig{
		Token:         token,
		SigningSecret: signingSecret,
	}
}
