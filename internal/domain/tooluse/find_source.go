package tooluse

import (
	"encoding/xml"
)

var FindSourceUsage = NewUsage("find_source", "Access PRIMARY SOURCE information from Slack conversations or GitHub discussions", []Parameter{
	NewParameter("uri", "The URI of the knowledge source (Slack threads or GitHub PRs) from document frontmatter. You can find it in the YAML frontmatter of the document. You must use the URI as it is, without any modifications.", true),
}, `<find_source>
<uri>https://app.slack.com/client/T01234567/C01234567/123456789.123456</uri>
</find_source>

IMPORTANT: This tool accesses PRIMARY SOURCE information:
- Use to retrieve original conversations that led to document creation
- Extract sources from document frontmatter using find_file first
- Provides raw context from original Slack threads or GitHub discussions
- Essential for understanding the full background of requirements
- More detailed than query_rag results, but limited to specific sources

Example patterns:
1. Retrieving Slack thread context: <find_source><uri>https://app.slack.com/client/T01234567/C01234567/T01234567-123456789.123456/234567890.234567</uri></find_source>
2. Accessing GitHub discussion: <find_source><uri>https://github.com/user/repo/pull/1</uri></find_source>`)

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
