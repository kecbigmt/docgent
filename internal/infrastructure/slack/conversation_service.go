package slack

import (
	"docgent/internal/application/port"
	"fmt"

	"github.com/slack-go/slack"
)

type ConversationService struct {
	slackAPI               *API
	channelID              string
	threadTimestamp        string
	sourceMessageTimestamp string
	userNameMap            map[string]string
}

func NewConversationService(slackAPI *API, channelID string, threadTimestamp string, sourceMessageTimestamp string) port.ConversationService {
	return &ConversationService{
		slackAPI:               slackAPI,
		channelID:              channelID,
		threadTimestamp:        threadTimestamp,
		sourceMessageTimestamp: sourceMessageTimestamp,
		userNameMap:            make(map[string]string),
	}
}

func (s *ConversationService) Reply(input string) error {
	slackClient := s.slackAPI.GetClient()

	slackClient.PostMessage(s.channelID, slack.MsgOptionText(input, false), slack.MsgOptionTS(s.threadTimestamp))

	return nil
}

func (s *ConversationService) GetHistory() ([]port.ConversationMessage, error) {
	client := s.slackAPI.GetClient()

	messages, _, _, err := client.GetConversationReplies(&slack.GetConversationRepliesParameters{
		ChannelID: s.channelID,
		Timestamp: s.threadTimestamp,
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
		Channel:   s.channelID,
		Timestamp: s.sourceMessageTimestamp,
	})
	if err != nil {
		return fmt.Errorf("failed to add eyes reaction: %w", err)
	}
	return nil
}

func (s *ConversationService) RemoveEyes() error {
	slackClient := s.slackAPI.GetClient()
	err := slackClient.RemoveReaction("eyes", slack.ItemRef{
		Channel:   s.channelID,
		Timestamp: s.sourceMessageTimestamp,
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
