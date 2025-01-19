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
	ToolParams ToolParams
}

type ToolParams []ToolParam

func (t ToolParams) GetOne(key string) (string, bool) {
	for _, param := range t {
		if param.Key == key {
			return param.Value, true
		}
	}
	return "", false
}

func (t ToolParams) GetAll(key string) []string {
	var values []string
	for _, param := range t {
		if param.Key == key {
			values = append(values, param.Value)
		}
	}
	return values
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

func (r Response) String() string {
	switch r.Type {
	case CompleteResponse:
		return fmt.Sprintf("<complete>%s</complete>", r.Message)
	case ErrorResponse:
		return fmt.Sprintf("<error>%s</error>", r.Message)
	case ToolUseResponse:
		var params []string
		for _, p := range r.ToolParams {
			params = append(params, fmt.Sprintf("<param:%s>%s</param:%s>", p.Key, p.Value, p.Key))
		}
		paramsStr := ""
		if len(params) > 0 {
			paramsStr = "\n" + strings.Join(params, "\n")
		}
		return fmt.Sprintf(`<tool_use:%s>
<message>%s</message>%s
</tool_use:%s>`, r.ToolType, r.Message, paramsStr, r.ToolType)
	default:
		return ""
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
