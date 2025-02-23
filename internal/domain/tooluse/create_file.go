package tooluse

import (
	"encoding/xml"
)

var CreateFileUsage = NewUsage("create_file", "Create a file", []Parameter{
	NewParameter("path", "The path to the file to create", true),
	NewParameter("content", "The content of the file to create", true),
	NewParameter("knowledge_source_uri", "The URIs of the knowledge sources (Slack threads or GitHub PRs)", true),
}, `<create_file>
<path>path/to/file.md</path>
<content>Hello, world!</content>
<knowledge_source_uri>https://slack.com/archives/C01234567/p123456789</knowledge_source_uri>
<knowledge_source_uri>https://github.com/user/repo/pull/1</knowledge_source_uri>
</create_file>`)

type CreateFile struct {
	XMLName             xml.Name `xml:"create_file"`
	Path                string   `xml:"path"`
	Content             string   `xml:"content"`
	KnowledgeSourceURIs []string `xml:"knowledge_source_uri"`
}

func (fc CreateFile) Match(cs ChangeFileCases) (string, bool, error) { return cs.CreateFile(fc) }

func NewCreateFile(path, content string, knowledgeSourceURIs []string) CreateFile {
	return CreateFile{
		XMLName:             xml.Name{Space: "", Local: "create_file"},
		Path:                path,
		Content:             content,
		KnowledgeSourceURIs: knowledgeSourceURIs,
	}
}
