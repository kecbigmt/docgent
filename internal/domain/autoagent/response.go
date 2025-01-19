package autoagent

import (
	"fmt"
)

type Response struct {
	Type       ResponseType
	Body       string
	ToolType   fmt.Stringer
	ToolParams interface{}
}

func (r Response) ToXMLString() string {
	str := "<response>"
	str += fmt.Sprintf("<type>%s</type>", r.Type)
	str += fmt.Sprintf("<body>%s</body>", r.Body)
	if r.ToolType != nil {
		str += fmt.Sprintf("<toolType>%s</toolType>", r.ToolType)
	}
	if r.ToolParams != nil {
		str += fmt.Sprintf("<toolParams>%s</toolParams>", r.ToolParams)
	}
	str += "</response>"
	return str
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
