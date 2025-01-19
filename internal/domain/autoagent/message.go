package autoagent

import "fmt"

type Message struct {
	Role    string
	Content string
}

type MessageHistory []Message

func (h MessageHistory) ToXMLString() string {
	str := "<history>"
	for _, message := range h {
		str += fmt.Sprintf("<item><role>%s</role><content>%s</content></item>", message.Role, message.Content)
	}
	str += "</history>"
	return str
}
