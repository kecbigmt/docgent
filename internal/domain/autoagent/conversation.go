package autoagent

type ConversationService interface {
	Reply(input string) error
}
