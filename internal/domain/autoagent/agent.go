package autoagent

import (
	"context"
	"fmt"

	"docgent-backend/internal/domain/autoagent/tooluse"
)

type Agent struct {
	chatModel           ChatModel
	conversationService ConversationService
	tools               tooluse.Cases
	systemInstruction   *SystemInstruction
}

func NewAgent(chatModel ChatModel, conversationService ConversationService, systemInstruction *SystemInstruction, tools tooluse.Cases) *Agent {
	return &Agent{chatModel: chatModel, conversationService: conversationService, tools: tools, systemInstruction: systemInstruction}
}

func (a *Agent) InitiateTaskLoop(ctx context.Context, task string, maxStepCount int) error {
	currentStepCount := 0
	nextMessage := NewMessage(UserRole, task)
	a.chatModel.SetSystemInstruction(a.systemInstruction.String())

	for currentStepCount <= maxStepCount {
		rawResponse, err := a.chatModel.SendMessage(ctx, nextMessage)
		if err != nil {
			go a.conversationService.Reply("Failed to generate response")
			return fmt.Errorf("failed to generate response: %w", err)
		}
		currentStepCount++

		toolUse, err := tooluse.Parse(rawResponse)
		if err != nil {
			go a.conversationService.Reply("Failed to parse response")
			return fmt.Errorf("failed to parse response: %w", err)
		}

		message, completed, err := toolUse.Match(a.tools)
		if err != nil {
			go a.conversationService.Reply("Failed to match tool use")
			return fmt.Errorf("failed to match tool use: %w", err)
		}
		if completed {
			go a.conversationService.Reply(message)
			return nil
		}
		nextMessage = NewMessage(UserRole, message)
	}

	go a.conversationService.Reply("Max task count reached")

	return nil
}
