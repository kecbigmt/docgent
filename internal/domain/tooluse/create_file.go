package tooluse

import (
	"encoding/xml"
)

var CreateFileUsage = NewUsage("create_file", "Create a file", []Parameter{
	NewParameter("path", "The path to the file to create", true),
	NewParameter("content", "The content of the file to create", true),
	NewParameter("conversation_uri", "The URI of the conversation that is the source of knowledge", true),
	NewParameter("proposal_uri", "The URI of the proposal that is the source of knowledge", false),
}, "<create_file><path>path/to/file.md</path><content>Hello, world!</content><conversation_uri>https://slack.com/archives/C01234567/p123456789</conversation_uri></create_file>")

type CreateFile struct {
	XMLName         xml.Name `xml:"create_file"`
	Path            string   `xml:"path"`
	Content         string   `xml:"content"`
	ConversationURI string   `xml:"conversation_uri"`
	ProposalURI     string   `xml:"proposal_uri,omitempty"`
}

func (fc CreateFile) Match(cs ChangeFileCases) (string, bool, error) { return cs.CreateFile(fc) }

func NewCreateFile(path, content string, conversationURI string, proposalURI string) CreateFile {
	return CreateFile{
		XMLName:         xml.Name{Space: "", Local: "create_file"},
		Path:            path,
		Content:         content,
		ConversationURI: conversationURI,
		ProposalURI:     proposalURI,
	}
}
