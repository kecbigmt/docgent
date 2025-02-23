package tooluse

import (
	"encoding/xml"
)

var AddKnowledgeSourcesUsage = NewUsage("add_knowledge_sources", "Add knowledge sources to an existing file", []Parameter{
	NewParameter("file_path", "The path to the file to add knowledge sources", true),
	NewParameter("uri", "The URIs of the knowledge sources (Slack threads or GitHub PRs)", true),
}, `<add_knowledge_sources>
<file_path>path/to/file.md</file_path>
<uri>https://slack.com/archives/C01234567/p123456789</uri>
<uri>https://github.com/user/repo/pull/1</uri>
</add_knowledge_sources>`)

type AddKnowledgeSources struct {
	XMLName  xml.Name `xml:"add_knowledge_sources"`
	FilePath string   `xml:"file_path"`
	URIs     []string `xml:"uri"`
}

func (aks AddKnowledgeSources) Match(cs Cases) (string, bool, error) {
	return cs.AddKnowledgeSources(aks)
}

func NewAddKnowledgeSources(filePath string, uris []string) AddKnowledgeSources {
	return AddKnowledgeSources{
		XMLName:  xml.Name{Space: "", Local: "add_knowledge_sources"},
		FilePath: filePath,
		URIs:     uris,
	}
}
