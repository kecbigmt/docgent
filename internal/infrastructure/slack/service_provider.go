package slack

import (
	"docgent/internal/application/port"
)

type ServiceProvider struct {
	slackAPI *API
}

func NewServiceProvider(slackAPI *API) *ServiceProvider {
	return &ServiceProvider{
		slackAPI: slackAPI,
	}
}

func (s *ServiceProvider) NewConversationService(uri *ConversationRef) port.ConversationService {
	return NewConversationService(s.slackAPI, uri)
}

func (s *ServiceProvider) NewSourceRepository() *SourceRepository {
	return NewSourceRepository(s.slackAPI)
}
