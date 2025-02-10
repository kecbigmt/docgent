package slack

import "docgent/internal/application/port"

type ServiceProvider struct {
	slackAPI *API
}

func NewServiceProvider(slackAPI *API) *ServiceProvider {
	return &ServiceProvider{
		slackAPI: slackAPI,
	}
}

func (s *ServiceProvider) NewConversationService(channelID, threadTimestamp, sourceMessageTimestamp string) port.ConversationService {
	return NewConversationService(s.slackAPI, channelID, threadTimestamp, sourceMessageTimestamp)
}
