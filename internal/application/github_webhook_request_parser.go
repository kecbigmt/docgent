package application

import "net/http"

type GitHubWebhookEvent interface {
	EventType() string
	InnerEvent() interface{}
}

type GitHubWebhookRequestParser interface {
	ParseRequest(r *http.Request) (GitHubWebhookEvent, error)
}
