package slack

import (
	application "docgent-backend/internal/application/slack"
	"docgent-backend/internal/domain"

	"github.com/slack-go/slack"
)

type ConversationService struct {
	slackAPI        application.API
	channelID       string
	threadTimestamp string
}

func NewConversationService(slackAPI application.API, channelID string, threadTimestamp string) domain.ConversationService {
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
