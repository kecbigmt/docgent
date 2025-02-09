package slack

import (
	"docgent-backend/internal/application/port"
	"fmt"

	"github.com/slack-go/slack"
)

type ConversationService struct {
	slackAPI               *API
	channelID              string
	threadTimestamp        string
	sourceMessageTimestamp string
}

func NewConversationService(slackAPI *API, channelID string, threadTimestamp string, sourceMessageTimestamp string) port.ConversationService {
	return &ConversationService{
		slackAPI:               slackAPI,
		channelID:              channelID,
		threadTimestamp:        threadTimestamp,
		sourceMessageTimestamp: sourceMessageTimestamp,
	}
}

func (s *ConversationService) Reply(input string) error {
	slackClient := s.slackAPI.GetClient()

	slackClient.PostMessage(s.channelID, slack.MsgOptionText(input, false), slack.MsgOptionTS(s.threadTimestamp))

	return nil
}

func (s *ConversationService) GetHistory() ([]port.ConversationMessage, error) {
	messages, _, _, err := s.slackAPI.GetClient().GetConversationReplies(&slack.GetConversationRepliesParameters{
		ChannelID: s.channelID,
		Timestamp: s.threadTimestamp,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get thread messages: %w", err)
	}

	conversationMessages := make([]port.ConversationMessage, 0, len(messages))
	for _, message := range messages {
		conversationMessages = append(conversationMessages, port.ConversationMessage{
			Author:  message.User,
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
