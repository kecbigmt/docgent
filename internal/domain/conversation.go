package domain

type ConversationService interface {
	Reply(input string) error
}
