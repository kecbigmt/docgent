package slack

import "docgent-backend/internal/application/port"

type ServiceProvider struct {
	slackAPI *API
}

func NewServiceProvider(slackAPI *API) *ServiceProvider {
	return &ServiceProvider{
		slackAPI: slackAPI,
	}
}

func (s *ServiceProvider) NewConversationService(channelID string, threadTimestamp string) port.ConversationService {
	return NewConversationService(s.slackAPI, channelID, threadTimestamp)
}
