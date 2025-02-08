package slack

import (
	"net/http"

	"github.com/slack-go/slack"
)

type API interface {
	GetClient() *slack.Client
	NewSecretsVerifier(http.Header) (slack.SecretsVerifier, error)
}
