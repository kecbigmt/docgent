package slack

import (
	"docgent/internal/application/port"
	"fmt"
	"regexp"

	"github.com/slack-go/slack"
)

type ConversationHandle struct {
	TeamID                 string
	ChannelID              string
	ThreadTimestamp        string
	SourceMessageTimestamp string
}

func NewConversationHandle(teamID, channelID, threadTimestamp, sourceMessageTimestamp string) ConversationHandle {
	return ConversationHandle{
		TeamID:                 teamID,
		ChannelID:              channelID,
		ThreadTimestamp:        threadTimestamp,
		SourceMessageTimestamp: sourceMessageTimestamp,
	}
}

type ConversationService struct {
	slackAPI    *API
	handle      ConversationHandle
	userNameMap map[string]string
}

func NewConversationService(slackAPI *API, handle ConversationHandle) port.ConversationService {
	return &ConversationService{
		slackAPI:    slackAPI,
		handle:      handle,
		userNameMap: make(map[string]string),
	}
}

// https://app.slack.com/client/{team_id}/{channel_id}/{thread_timestamp}
// or https://app.slack.com/client/{team_id}/{channel_id}/thread/{channel_id}-{parent_thread_timestamp}/{child_thread_timestamp}
var re1 = regexp.MustCompile(`^https://app\.slack\.com/client/([^/]+)/([^/]+)/([^/]+)$`)
var re2 = regexp.MustCompile(`^https://app\.slack\.com/client/([^/]+)/([^/]+)/thread/([^/]+)-([^/]+)/([^/]+)$`)

func ParseConversationURI(uri string) (ConversationHandle, error) {
	matches := re1.FindStringSubmatch(uri)
	if len(matches) == 4 {
		return ConversationHandle{
			TeamID:                 matches[1],
			ChannelID:              matches[2],
			ThreadTimestamp:        matches[3],
			SourceMessageTimestamp: matches[3],
		}, nil
	}
	matches2 := re2.FindStringSubmatch(uri)
	if len(matches2) == 6 {
		return ConversationHandle{
			TeamID:                 matches2[1],
			ChannelID:              matches2[2],
			ThreadTimestamp:        matches2[4],
			SourceMessageTimestamp: matches2[5],
		}, nil
	}
	return ConversationHandle{}, fmt.Errorf("invalid URI: %s", uri)
}

func (s *ConversationService) Reply(input string) error {
	slackClient := s.slackAPI.GetClient()

	slackClient.PostMessage(s.handle.ChannelID, slack.MsgOptionText(input, false), slack.MsgOptionTS(s.handle.ThreadTimestamp))

	return nil
}

func (s *ConversationService) GetURI() string {
	if s.handle.ThreadTimestamp == s.handle.SourceMessageTimestamp {
		return fmt.Sprintf("https://app.slack.com/client/%s/%s/%s", s.handle.TeamID, s.handle.ChannelID, s.handle.ThreadTimestamp)
	}
	return fmt.Sprintf("https://app.slack.com/client/%s/%s/thread/%s-%s/%s", s.handle.TeamID, s.handle.ChannelID, s.handle.ChannelID, s.handle.ThreadTimestamp, s.handle.SourceMessageTimestamp)
}

func (s *ConversationService) GetHistory() ([]port.ConversationMessage, error) {
	client := s.slackAPI.GetClient()

	messages, _, _, err := client.GetConversationReplies(&slack.GetConversationRepliesParameters{
		ChannelID: s.handle.ChannelID,
		Timestamp: s.handle.ThreadTimestamp,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get thread messages: %w", err)
	}

	conversationMessages := make([]port.ConversationMessage, 0, len(messages))
	for _, message := range messages {
		author, err := s.getAuthorName(&message)
		if err != nil {
			return nil, err
		}

		conversationMessages = append(conversationMessages, port.ConversationMessage{
			Author:  author,
			Content: message.Text,
		})
	}

	return conversationMessages, nil
}

func (s *ConversationService) MarkEyes() error {
	slackClient := s.slackAPI.GetClient()
	err := slackClient.AddReaction("eyes", slack.ItemRef{
		Channel:   s.handle.ChannelID,
		Timestamp: s.handle.SourceMessageTimestamp,
	})
	if err != nil {
		return fmt.Errorf("failed to add eyes reaction: %w", err)
	}
	return nil
}

func (s *ConversationService) RemoveEyes() error {
	slackClient := s.slackAPI.GetClient()
	err := slackClient.RemoveReaction("eyes", slack.ItemRef{
		Channel:   s.handle.ChannelID,
		Timestamp: s.handle.SourceMessageTimestamp,
	})
	if err != nil {
		return fmt.Errorf("failed to remove eyes reaction: %w", err)
	}
	return nil
}

func (s *ConversationService) getAuthorName(message *slack.Message) (string, error) {
	// Username is only available in bot messages
	if message.Username != "" {
		return message.Username, nil
	}

	// if it's not a bot message, use the user name cache
	_, exists := s.userNameMap[message.User]
	if exists {
		return s.userNameMap[message.User], nil
	}

	// if the user name is not in the cache, get the user info
	userInfo, err := s.slackAPI.GetClient().GetUserInfo(message.User)
	if err != nil {
		return "", fmt.Errorf("failed to get user info: %w", err)
	}

	// if the display name is set, use it
	if userInfo.Profile.DisplayName != "" {
		return userInfo.Profile.DisplayName, nil
	}

	// if the real name is set, use it
	return userInfo.Profile.RealName, nil
}
