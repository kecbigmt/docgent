package autoagent

import (
	"context"
	"fmt"

	"docgent-backend/internal/domain/autoagent/tooluse"
)

type Agent struct {
	chatModel         ChatModel
	tools             tooluse.Cases
	systemInstruction *SystemInstruction
}

func NewAgent(chatModel ChatModel, systemInstruction *SystemInstruction, tools tooluse.Cases) *Agent {
	return &Agent{chatModel: chatModel, tools: tools, systemInstruction: systemInstruction}
}

func (a *Agent) InitiateTaskLoop(ctx context.Context, task string, maxStepCount int) error {
	currentStepCount := 0
	nextMessage := NewMessage(UserRole, task)
	a.chatModel.SetSystemInstruction(a.systemInstruction.String())

	for currentStepCount <= maxStepCount {
		rawResponse, err := a.chatModel.SendMessage(ctx, nextMessage)
		if err != nil {
			return fmt.Errorf("failed to generate response: %w", err)
		}
		currentStepCount++

		toolUse, err := tooluse.Parse(rawResponse)
		if err != nil {
			return fmt.Errorf("failed to parse response: %w", err)
		}

		message, completed, err := toolUse.Match(a.tools)
		if err != nil {
			return fmt.Errorf("failed to match tool use: %w", err)
		}
		if completed {
			return nil
		}
		nextMessage = NewMessage(UserRole, message)
	}

	return fmt.Errorf("max task count reached")
}

	go a.conversationService.Reply("Max task count reached")

	return nil
}
