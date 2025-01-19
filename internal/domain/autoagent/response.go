package autoagent

import (
	"fmt"
	"regexp"
	"strings"
)

type Response struct {
	Type       ResponseType
	Message    string
	ToolType   string
	ToolParams []ToolParam
}

type ToolParam struct {
	Key   string
	Value string
}

type ResponseType int

const (
	ToolUseResponse ResponseType = iota
	CompleteResponse
	ErrorResponse
)

func (t ResponseType) String() string {
	switch t {
	case ToolUseResponse:
		return "tool_use"
	case CompleteResponse:
		return "complete"
	case ErrorResponse:
		return "error"
	default:
		return "unknown"
	}
}

func ParseResponse(raw string) (Response, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return Response{}, fmt.Errorf("empty response")
	}

	// Complete response
	if match := regexp.MustCompile(`^<complete>(.*)</complete>$`).FindStringSubmatch(raw); len(match) == 2 {
		return Response{
			Type:    CompleteResponse,
			Message: match[1],
		}, nil
	}

	// Error response
	if match := regexp.MustCompile(`^<error>(.*)</error>$`).FindStringSubmatch(raw); len(match) == 2 {
		return Response{
			Type:    ErrorResponse,
			Message: match[1],
		}, nil
	}

	// Tool use response
	toolUsePattern := regexp.MustCompile(`(?s)^<tool_use:([^>]+)>\s*<message>(.*?)</message>\s*(.*?)</tool_use:[^>]+>$`)
	if match := toolUsePattern.FindStringSubmatch(raw); len(match) == 4 {
		toolName := match[1]
		message := match[2]
		paramsRaw := match[3]

		// Parse parameters
		var params []ToolParam
		paramPattern := regexp.MustCompile(`<param:([^>]+)>(.*?)</param:[^>]+>`)
		paramMatches := paramPattern.FindAllStringSubmatch(paramsRaw, -1)
		for _, paramMatch := range paramMatches {
			params = append(params, ToolParam{
				Key:   paramMatch[1],
				Value: paramMatch[2],
			})
		}

		return Response{
			Type:       ToolUseResponse,
			Message:    message,
			ToolType:   toolName,
			ToolParams: params,
		}, nil
	}

	return Response{}, fmt.Errorf("invalid response format")
}
