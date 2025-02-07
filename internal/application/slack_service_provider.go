package application

import "docgent-backend/internal/domain"

type SlackServiceProvider interface {
	NewConversationService(channelID string, threadTimestamp string) domain.ConversationService
}
