package tooluse

import (
	"encoding/xml"
)

var CreateFileUsage = NewUsage("create_file", "Create a file", []Parameter{
	NewParameter("path", "The path to the file to create", true),
	NewParameter("content", "The content of the file to create", true),
	NewParameter("source_uri", "The URIs of the knowledge sources (Slack threads or GitHub PRs)", true),
}, `<create_file>
<path>path/to/file.md</path>
<content>Hello, world!</content>
<source_uri>https://slack.com/archives/C01234567/p123456789</source_uri>
<source_uri>https://github.com/user/repo/pull/1</source_uri>
</create_file>`)

type CreateFile struct {
	XMLName    xml.Name `xml:"create_file"`
	Path       string   `xml:"path"`
	Content    string   `xml:"content"`
	SourceURIs []string `xml:"source_uri"`
}

func (fc CreateFile) Match(cs ChangeFileCases) (string, bool, error) { return cs.CreateFile(fc) }

func NewCreateFile(path, content string, sourceURIs []string) CreateFile {
	return CreateFile{
		XMLName:    xml.Name{Space: "", Local: "create_file"},
		Path:       path,
		Content:    content,
		SourceURIs: sourceURIs,
	}
}
