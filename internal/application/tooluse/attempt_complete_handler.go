package tooluse

import (
	"fmt"

	"docgent/internal/application/port"
	"docgent/internal/domain/tooluse"
)

// AttemptCompleteHandler は attempt_complete ツールのハンドラーです
type AttemptCompleteHandler struct {
	conversationService port.ConversationService
}

func NewAttemptCompleteHandler(conversationService port.ConversationService) *AttemptCompleteHandler {
	return &AttemptCompleteHandler{
		conversationService: conversationService,
	}
}

func (h *AttemptCompleteHandler) Handle(toolUse tooluse.AttemptComplete) (string, bool, error) {
	if err := h.conversationService.Reply(toolUse.Message); err != nil {
		return "", false, fmt.Errorf("failed to reply: %w", err)
	}
	return "", true, nil
}
