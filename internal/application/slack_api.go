package application

import (
	"net/http"

	"github.com/slack-go/slack"
)

type SlackAPI interface {
	GetClient() *slack.Client
	NewSecretsVerifier(http.Header) (slack.SecretsVerifier, error)
}
