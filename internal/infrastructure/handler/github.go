package handler

type GitHubEventRoute interface {
	ConsumeEvent(event interface{})
	EventType() string
}
