package slack

import (
	"docgent-backend/internal/domain"
	"fmt"

	"github.com/slack-go/slack"
)

type ConversationService struct {
	slackAPI        *API
	channelID       string
	threadTimestamp string
}

func NewConversationService(slackAPI *API, channelID string, threadTimestamp string) domain.ConversationService {
	return &ConversationService{
		slackAPI:        slackAPI,
		channelID:       channelID,
		threadTimestamp: threadTimestamp,
	}
}

func (s *ConversationService) Reply(input string) error {
	slackClient := s.slackAPI.GetClient()

	slackClient.PostMessage(s.channelID, slack.MsgOptionText(input, false), slack.MsgOptionTS(s.threadTimestamp))

	return nil
}

func (s *ConversationService) GetHistory() ([]domain.ConversationMessage, error) {
	messages, _, _, err := s.slackAPI.GetClient().GetConversationReplies(&slack.GetConversationRepliesParameters{
		ChannelID: s.channelID,
		Timestamp: s.threadTimestamp,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get thread messages: %w", err)
	}

	conversationMessages := make([]domain.ConversationMessage, 0, len(messages))
	for _, message := range messages {
		conversationMessages = append(conversationMessages, domain.ConversationMessage{
			Author:  message.User,
			Content: message.Text,
		})
	}

	return conversationMessages, nil
}
