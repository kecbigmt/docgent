package github

import (
	"fmt"
	"strings"

	"docgent/internal/application/port"
	"docgent/internal/domain/tooluse"
)

// ResponseFormatter implements the port.ResponseFormatter interface for GitHub
type ResponseFormatter struct{}

func NewResponseFormatter() port.ResponseFormatter {
	return &ResponseFormatter{}
}

func (f *ResponseFormatter) FormatResponse(toolUse tooluse.AttemptComplete) (string, error) {
	var builder strings.Builder

	// Format the messages - GitHub uses Markdown format
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

	// Format the sources - GitHub supports Markdown links
	if len(toolUse.Sources) > 0 {
		builder.WriteString("\n")
		for _, s := range toolUse.Sources {
			// TODO: Update to use GitHub permalinks when available
			builder.WriteString(fmt.Sprintf("[^%s]: %s\n", s.ID, s.URI))
		}
	}

	return strings.TrimSpace(builder.String()), nil
}
