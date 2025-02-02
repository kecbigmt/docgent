package tooluse

import "encoding/xml"

var AttemptCompleteUsage = NewUsage("attempt_complete", "You should use this tool when you think you have completed the task.", []Parameter{
	NewParameter("message", "Let the user know what you have done.", true),
}, "<attempt_complete><message>Completed the task</message></attempt_complete>")

type AttemptComplete struct {
	XMLName xml.Name `xml:"attempt_complete"`
	Message string   `xml:"message"`
}

func NewAttemptComplete(message string) AttemptComplete {
	return AttemptComplete{
		XMLName: xml.Name{Local: "attempt_complete"},
		Message: message,
	}
}

func (ac AttemptComplete) Match(cs Cases) (string, bool, error) { return cs.AttemptComplete(ac) }
