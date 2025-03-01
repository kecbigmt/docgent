package port

import (
	"docgent/internal/domain/tooluse"
)

// ResponseFormatter is an interface responsible for formatting response messages
// Each platform (Slack, GitHub, etc.) implements its own specific formatting
type ResponseFormatter interface {
	// FormatResponse formats the result of the AttemptComplete tool into a platform-specific format
	FormatResponse(toolUse tooluse.AttemptComplete) (string, error)
}
