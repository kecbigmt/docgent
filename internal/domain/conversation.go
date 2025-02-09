package domain

type ConversationService interface {
	Reply(input string) error
	GetHistory() ([]ConversationMessage, error)
}

type ConversationMessage struct {
	Author  string
	Content string
}
