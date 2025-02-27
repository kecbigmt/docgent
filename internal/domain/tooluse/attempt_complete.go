package tooluse

import (
	"encoding/xml"
	"strings"
)

var AttemptCompleteUsage = NewUsage("attempt_complete", "You should use this tool only when you think you have completed the task.", []Parameter{
	NewParameter("message", "Let the user know what you have done. You can include one or more <message> tags to describe what you have done. If you used any sources, you should indicate which messages correspond to which sources by adding numbers separated by commas to the `source` attribute of the <message> tags.", true),
	NewParameter("source", "The source names you used to complete the task. `id` attribute should correspond to the `source` attribute of the <message> tags. `uri` attribute is the URI of the source.", false),
}, `Simple example:

<attempt_complete>
<message>Here is the answer:
- Docgent is a agent that can help you with your documentation.
- Docgent can create documents based on chat history.</message>
</attempt_complete>

Example with sources:
<attempt_complete>
<message>Here is the answer:
</message>
<message source="1,2">- Docgent is a agent that can help you with your documentation</message>
<message source="2">- Docgent can create documents based on chat history.</message>
</attempt_complete>
<source id="1" uri="https://github.com/owner/repo/blob/a1b2c3d4e5f6g7h8i9j0k1l2m3n4o5p6q7r8s9t0/docs/what-is-docgent.md">What is Docgent?</source>
<source id="2" uri="https://github.com/owner/repo/blob/a1b2c3d4e5f6g7h8i9j0k1l2m3n4o5p6q7r8s9t0/docs/docgent-features.md">Docgent Features</source>
</attempt_complete>`)

type AttemptComplete struct {
	XMLName  xml.Name  `xml:"attempt_complete"`
	Messages []Message `xml:"message"`
	Sources  []Source  `xml:"source"`
}

func NewAttemptComplete(messages []Message, sources []Source) AttemptComplete {
	return AttemptComplete{
		XMLName:  xml.Name{Local: "attempt_complete"},
		Messages: messages,
		Sources:  sources,
	}
}

type Message struct {
	XMLName  xml.Name `xml:"message"`
	SourceID string   `xml:"source,attr,omitempty"`
	Text     string   `xml:",chardata"`
}

// GetSourceIDs はカンマ区切りのSourceIDをスライスに変換して返す
func (m Message) GetSourceIDs() []string {
	if m.SourceID == "" {
		return nil
	}
	return strings.Split(m.SourceID, ",")
}

func NewMessage(text string) Message {
	return Message{
		XMLName: xml.Name{Local: "message"},
		Text:    text,
	}
}

func NewMessageWithSourceID(text string, sourceID string) Message {
	return Message{
		XMLName:  xml.Name{Local: "message"},
		Text:     text,
		SourceID: sourceID,
	}
}

// NewMessageWithSourceIDs は複数のソースIDを持つメッセージを作成する
func NewMessageWithSourceIDs(text string, sourceIDs []string) Message {
	return Message{
		XMLName:  xml.Name{Local: "message"},
		Text:     text,
		SourceID: strings.Join(sourceIDs, ","),
	}
}

type Source struct {
	XMLName xml.Name `xml:"source"`
	ID      string   `xml:"id,attr"`
	URI     string   `xml:"uri,attr"`
	Name    string   `xml:",chardata"`
}

func NewSource(id string, uri string, name string) Source {
	return Source{
		XMLName: xml.Name{Local: "source"},
		ID:      id,
		URI:     uri,
		Name:    name,
	}
}

func (ac AttemptComplete) Match(cs Cases) (string, bool, error) { return cs.AttemptComplete(ac) }
