package tooluse

import (
	"fmt"
	"strings"

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
	var builder strings.Builder
	for _, m := range toolUse.Messages {
		builder.WriteString(m.Text)
		if m.SourceID != "" {
			sourceIDs := m.GetSourceIDs()
			for _, sourceID := range sourceIDs {
				builder.WriteString(fmt.Sprintf("[^%s]", sourceID))
			}
		}
		builder.WriteString("\n")
	}
	if len(toolUse.Sources) > 0 {
		builder.WriteString("\n")
		for _, s := range toolUse.Sources {
			// TODO: 現状のURIはURLとして利用可能な状態ではないため、そのまま表示。あとで修正する
			// builder.WriteString(fmt.Sprintf("[^%s]: <%s|%s>\n", s.ID, s.URI, s.Name))
			builder.WriteString(fmt.Sprintf("[^%s]: %s\n", s.ID, s.URI))
		}
	}
	message := strings.TrimSpace(builder.String())

	if err := h.conversationService.Reply(message, true); err != nil {
		return "", false, fmt.Errorf("failed to reply: %w", err)
	}
	return "", true, nil
}
