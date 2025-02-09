package port

type ConversationService interface {
	Reply(input string) error
	GetHistory() ([]ConversationMessage, error)
	MarkEyes() error
	RemoveEyes() error
}

type ConversationMessage struct {
	Author  string
	Content string
}
