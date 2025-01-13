package application

import "github.com/slack-go/slack/slackevents"

type GitHubAppParams struct {
	Owner          string
	Repo           string
	BaseBranch     string
	InstallationID int64
}

type SlackEventRoute interface {
	ConsumeEvent(event slackevents.EventsAPIInnerEvent, githubAppParams GitHubAppParams)
	EventType() string
}
