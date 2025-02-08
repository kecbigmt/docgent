package slack

import (
	"docgent-backend/internal/application"
	"docgent-backend/internal/domain"
)

type ServiceProvider struct {
	slackAPI application.SlackAPI
}

func NewServiceProvider(slackAPI application.SlackAPI) application.SlackServiceProvider {
	return &ServiceProvider{
		slackAPI: slackAPI,
	}
}

func (s *ServiceProvider) NewConversationService(channelID string, threadTimestamp string) domain.ConversationService {
	return NewConversationService(s.slackAPI, channelID, threadTimestamp)
}
