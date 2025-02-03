package domain

import (
	"context"
)

type ChatModel interface {
	SetSystemInstruction(instruction string) error
	SendMessage(ctx context.Context, message Message) (string, error)
	GetHistory() ([]Message, error)
}
