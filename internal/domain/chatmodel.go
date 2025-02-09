package domain

import (
	"context"
)

type ChatModel interface {
	StartChat(systemInstruction string) ChatSession
}

type ChatSession interface {
	SendMessage(ctx context.Context, message string) (string, error)
	GetHistory() ([]Message, error)
}
