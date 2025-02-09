package slack

import (
	"errors"
	"fmt"
	"io"
	"net/http"

	"github.com/slack-go/slack/slackevents"
	"go.uber.org/zap"
)

var (
	ErrInvalidHeader       = errors.New("invalid header")
	ErrInvalidBody         = errors.New("invalid body")
	ErrUnauthorizedRequest = errors.New("unauthorized request")
	ErrInvalidEvent        = errors.New("invalid event")
)

type WebhookRequestParser struct {
	api    *API
	logger *zap.Logger
}

func NewWebhookRequestParser(api *API, logger *zap.Logger) *WebhookRequestParser {
	return &WebhookRequestParser{api: api, logger: logger}
}

func (p *WebhookRequestParser) ParseRequest(r *http.Request) (slackevents.EventsAPIEvent, error) {
	sv, err := p.api.NewSecretsVerifier(r.Header)
	if err != nil {
		p.logger.Error("failed to create secrets verifier", zap.Error(err))
		return slackevents.EventsAPIEvent{}, fmt.Errorf("failed to create secrets verifier: %w", ErrInvalidHeader)
	}

	bodyReader := io.TeeReader(r.Body, &sv)
	body, err := io.ReadAll(bodyReader)
	if err != nil {
		p.logger.Error("failed to read request", zap.Error(err))
		return slackevents.EventsAPIEvent{}, fmt.Errorf("failed to read request: %w", ErrInvalidBody)
	}

	if err := sv.Ensure(); err != nil {
		p.logger.Error("failed to verify request", zap.Error(err))
		return slackevents.EventsAPIEvent{}, fmt.Errorf("failed to verify request: %w", ErrUnauthorizedRequest)
	}

	event, err := slackevents.ParseEvent(body, slackevents.OptionNoVerifyToken())
	if err != nil {
		p.logger.Error("failed to parse event", zap.Error(err))
		return slackevents.EventsAPIEvent{}, fmt.Errorf("failed to parse event: %w", ErrInvalidEvent)
	}

	return event, nil
}
