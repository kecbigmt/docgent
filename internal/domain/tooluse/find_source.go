package tooluse

import (
	"encoding/xml"
)

var FindSourceUsage = NewUsage("find_source", "Find a knowledge source", []Parameter{
	NewParameter("uri", "The URI of the knowledge source (Slack threads or GitHub PRs)", true),
}, `<find_source>
<uri>https://app.slack.com/client/T00000000/C00000000/thread/T00000000-00000000</uri>
</find_source>`)

type FindSource struct {
	XMLName xml.Name `xml:"find_source"`
	URI     string   `xml:"uri"`
}

func (fs FindSource) Match(cs Cases) (string, bool, error) {
	return cs.FindSource(fs)
}

func NewFindSource(uri string) FindSource {
	return FindSource{
		XMLName: xml.Name{Space: "", Local: "find_source"},
		URI:     uri,
	}
}
