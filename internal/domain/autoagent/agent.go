package autoagent

import "context"

type Agent interface {
	SetSystemInstruction(instruction string) error
	SendMessage(ctx context.Context, message Message) (Response, error)
	GetHistory() ([]Message, error)
}
