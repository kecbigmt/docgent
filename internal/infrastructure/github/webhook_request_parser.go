package github

import (
	"fmt"
	"net/http"

	"github.com/google/go-github/v68/github"
)

/**
 * WebhookEvent
 */

type WebhookEvent struct {
	eventType  string
	innerEvent interface{}
}

func NewWebhookEvent(eventType string, innerEvent interface{}) *WebhookEvent {
	return &WebhookEvent{eventType: eventType, innerEvent: innerEvent}
}

func (e *WebhookEvent) EventType() string {
	return e.eventType
}

func (e *WebhookEvent) InnerEvent() interface{} {
	return e.innerEvent
}

/**
 * WebhookRequestParser
 */

type WebhookRequestParser struct {
	webhookSecret string
}

func NewWebhookRequestParser(webhookSecret string) *WebhookRequestParser {
	return &WebhookRequestParser{webhookSecret: webhookSecret}
}

func (p *WebhookRequestParser) ParseRequest(r *http.Request) (*WebhookEvent, error) {
	payload, err := github.ValidatePayload(r, []byte(p.webhookSecret))
	if err != nil {
		return nil, fmt.Errorf("failed to validate payload: %w", err)
	}

	ev, err := github.ParseWebHook(github.WebHookType(r), payload)
	if err != nil {
		return nil, fmt.Errorf("failed to parse webhook: %w", err)
	}

	switch event := ev.(type) {
	case *github.IssueCommentEvent:
		return NewWebhookEvent("issue_comment", event), nil
	}

	return nil, fmt.Errorf("unsupported event type: %T", ev)
}
