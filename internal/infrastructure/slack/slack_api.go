package slack

import (
	"net/http"

	slacklib "github.com/slack-go/slack"
)

type API struct {
	client        *slacklib.Client
	signingSecret string
}

func NewAPI(token, signingSecret string) *API {
	return &API{client: slacklib.New(token), signingSecret: signingSecret}
}

func (a *API) GetClient() *slacklib.Client {
	return a.client
}

func (a *API) NewSecretsVerifier(header http.Header) (slacklib.SecretsVerifier, error) {
	return slacklib.NewSecretsVerifier(header, a.signingSecret)
}
