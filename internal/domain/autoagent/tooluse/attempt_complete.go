package tooluse

import "encoding/xml"

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
