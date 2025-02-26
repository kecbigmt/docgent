package port

import (
	"docgent/internal/domain/data"
	"encoding/xml"
)

type ConversationService interface {
	Reply(input string, withMention bool) error
	GetHistory() (ConversationHistory, error)
	URI() *data.URI
	MarkEyes() error
	RemoveEyes() error
}

type ConversationHistory struct {
	URI      *data.URI
	Messages []ConversationMessage
}

type ConversationMessage struct {
	Author       string
	Content      string
	YouMentioned bool
	IsYou        bool
}

func (c ConversationHistory) ToXML() string {
	messages := make([]conversationMessage, len(c.Messages))
	for i, message := range c.Messages {
		messages[i] = conversationMessage{
			Author:       message.Author,
			Content:      message.Content,
			YouMentioned: message.YouMentioned,
			IsYou:        message.IsYou,
		}
	}
	history := conversationHistory{
		URI:      c.URI.String(),
		Messages: messages,
	}

	xmlData, err := xml.MarshalIndent(history, "", "  ")
	if err != nil {
		return ""
	}
	return string(xmlData)
}

type conversationHistory struct {
	XMLName  xml.Name              `xml:"conversation"`
	URI      string                `xml:"uri,attr"`
	Messages []conversationMessage `xml:"message"`
}

type conversationMessage struct {
	XMLName      xml.Name `xml:"message"`
	Author       string   `xml:"author,attr"`
	IsYou        bool     `xml:"is_you,attr,omitempty"`
	YouMentioned bool     `xml:"you_mentioned,attr,omitempty"`
	Content      string   `xml:",chardata"`
}
