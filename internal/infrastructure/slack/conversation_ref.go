package slack

import (
	"docgent/internal/domain/data"
	"fmt"
	"regexp"
)

type ConversationRef struct {
	teamID                 string
	channelID              string
	threadTimestamp        string
	sourceMessageTimestamp string
}

func NewConversationRef(teamID, channelID, threadTimestamp, sourceMessageTimestamp string) *ConversationRef {
	return &ConversationRef{teamID: teamID, channelID: channelID, threadTimestamp: threadTimestamp, sourceMessageTimestamp: sourceMessageTimestamp}
}

func (u *ConversationRef) TeamID() string {
	return u.teamID
}

func (u *ConversationRef) ChannelID() string {
	return u.channelID
}

func (u *ConversationRef) ThreadTimestamp() string {
	return u.threadTimestamp
}

func (u *ConversationRef) SourceMessageTimestamp() string {
	return u.sourceMessageTimestamp
}

func (u *ConversationRef) ToURI() data.URI {
	if u.threadTimestamp == u.sourceMessageTimestamp {
		return data.NewURIUnsafe(fmt.Sprintf("https://app.slack.com/client/%s/%s/%s", u.teamID, u.channelID, u.threadTimestamp))
	}
	return data.NewURIUnsafe(fmt.Sprintf("https://app.slack.com/client/%s/%s/thread/%s-%s/%s", u.teamID, u.channelID, u.channelID, u.threadTimestamp, u.sourceMessageTimestamp))
}

// https://app.slack.com/client/{team_id}/{channel_id}/{thread_timestamp}
// or https://app.slack.com/client/{team_id}/{channel_id}/thread/{channel_id}-{parent_thread_timestamp}/{child_thread_timestamp}
var re1 = regexp.MustCompile(`^https://app\.slack\.com/client/([^/]+)/([^/]+)/([^/]+)$`)
var re2 = regexp.MustCompile(`^https://app\.slack\.com/client/([^/]+)/([^/]+)/thread/([^/]+)-([^/]+)/([^/]+)$`)

func ParseConversationRef(uri *data.URI) (*ConversationRef, error) {
	rawURI := uri.Value()
	matches := re1.FindStringSubmatch(rawURI)
	if len(matches) == 4 {
		return NewConversationRef(matches[1], matches[2], matches[3], matches[3]), nil
	}
	matches2 := re2.FindStringSubmatch(rawURI)
	if len(matches2) == 6 {
		return NewConversationRef(matches2[1], matches2[2], matches2[4], matches2[5]), nil
	}
	return nil, fmt.Errorf("invalid URI: %s", uri)
}
