package slack

import (
	application "docgent-backend/internal/application/slack"
	"docgent-backend/internal/domain"
)

type ServiceProvider struct {
	slackAPI application.API
}

func NewServiceProvider(slackAPI application.API) application.ServiceProvider {
	return &ServiceProvider{
		slackAPI: slackAPI,
	}
}

func (s *ServiceProvider) NewConversationService(channelID string, threadTimestamp string) domain.ConversationService {
	return NewConversationService(s.slackAPI, channelID, threadTimestamp)
}
