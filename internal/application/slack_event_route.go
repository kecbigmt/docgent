package application

import "github.com/slack-go/slack/slackevents"

type SlackEventRoute interface {
	ConsumeEvent(event slackevents.EventsAPIInnerEvent, workspace Workspace)
	EventType() string
}
