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

func (s *ServiceProvider) NewConversationService(ref *ConversationRef, fromUserID string) port.ConversationService {
	return NewConversationService(s.slackAPI, ref, fromUserID)
}

func (s *ServiceProvider) NewSourceRepository() *SourceRepository {
	return NewSourceRepository(s.slackAPI)
}

func (s *ServiceProvider) NewResponseFormatter() port.ResponseFormatter {
	return NewResponseFormatter()
}
