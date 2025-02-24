package port

import "docgent/internal/domain/data"

type ConversationService interface {
	Reply(input string) error
	GetHistory() ([]ConversationMessage, error)
	URI() data.URI
	MarkEyes() error
	RemoveEyes() error
}

type ConversationMessage struct {
	Author  string
	Content string
}
