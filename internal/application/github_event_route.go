package application

type GitHubEventRoute interface {
	ConsumeEvent(event interface{})
	EventType() string
}
