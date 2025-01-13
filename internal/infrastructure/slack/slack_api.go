package slack

import (
	"net/http"

	"github.com/slack-go/slack"
)

type API struct {
	client        *slack.Client
	signingSecret string
}

func NewAPI(token, signingSecret string) *API {
	return &API{client: slack.New(token), signingSecret: signingSecret}
}

func (a *API) GetClient() *slack.Client {
	return a.client
}

func (a *API) NewSecretsVerifier(header http.Header) (slack.SecretsVerifier, error) {
	return slack.NewSecretsVerifier(header, a.signingSecret)
}
