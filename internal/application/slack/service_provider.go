package slack

import "docgent-backend/internal/domain"

type ServiceProvider interface {
	NewConversationService(channelID string, threadTimestamp string) domain.ConversationService
}
