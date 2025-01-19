package autoagent

import (
	"fmt"
)

type Response struct {
	Type       ResponseType `json:"type"`
	Message    string       `json:"message"`
	ToolType   string       `json:"toolType,omitempty"`
	ToolParams ToolParams   `json:"toolParams,omitempty"`
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
	Key   string `json:"k"`
	Value string `json:"v"`
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

func (t ResponseType) MarshalJSON() ([]byte, error) {
	return []byte(fmt.Sprintf(`"%s"`, t.String())), nil
}

func (t *ResponseType) UnmarshalJSON(data []byte) error {
	switch string(data) {
	case `"tool_use"`:
		*t = ToolUseResponse
	case `"complete"`:
		*t = CompleteResponse
	case `"error"`:
		*t = ErrorResponse
	default:
		return fmt.Errorf("invalid response type")
	}
	return nil
}
