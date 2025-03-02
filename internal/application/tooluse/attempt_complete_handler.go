package tooluse

import (
	"fmt"

	"docgent/internal/application/port"
	"docgent/internal/domain/tooluse"
)

// AttemptCompleteHandler handles the attempt_complete tool
type AttemptCompleteHandler struct {
	conversationService port.ConversationService
	responseFormatter   port.ResponseFormatter
}

func NewAttemptCompleteHandler(
	conversationService port.ConversationService,
	responseFormatter port.ResponseFormatter,
) *AttemptCompleteHandler {
	return &AttemptCompleteHandler{
		conversationService: conversationService,
		responseFormatter:   responseFormatter,
	}
}

func (h *AttemptCompleteHandler) Handle(toolUse tooluse.AttemptComplete) (string, bool, error) {
	// Use the formatter to get the platform-specific formatted message
	message, err := h.responseFormatter.FormatResponse(toolUse)
	if err != nil {
		return "", false, fmt.Errorf("failed to format response: %w", err)
	}

	// Send the formatted message using the conversation service
	if err := h.conversationService.Reply(message, true); err != nil {
		return "", false, fmt.Errorf("failed to reply: %w", err)
	}

	return "", true, nil
}
