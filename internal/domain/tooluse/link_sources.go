package tooluse

import (
	"encoding/xml"
)

var LinkSourcesUsage = NewUsage("link_sources", "Link knowledge sources to an existing file - CRITICAL for preserving context", []Parameter{
	NewParameter("file_path", "The path to the file to link knowledge sources", true),
	NewParameter("uri", "The URIs of the knowledge sources (Slack threads, GitHub PRs, etc.). You can find them in the <conversation> tags.", true),
}, `<link_sources>
<file_path>path/to/file.md</file_path>
<uri>https://app.slack.com/client/T00000000/C00000000/thread/T00000000-00000000</uri>
<uri>https://github.com/user/repo/pull/1</uri>
</link_sources>`)

type LinkSources struct {
	XMLName  xml.Name `xml:"link_sources"`
	FilePath string   `xml:"file_path"`
	URIs     []string `xml:"uri"`
}

func (ls LinkSources) Match(cs Cases) (string, bool, error) {
	return cs.LinkSources(ls)
}

func NewLinkSources(filePath string, uris []string) LinkSources {
	return LinkSources{
		XMLName:  xml.Name{Space: "", Local: "link_sources"},
		FilePath: filePath,
		URIs:     uris,
	}
}
