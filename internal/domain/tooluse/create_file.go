package tooluse

import (
	"encoding/xml"
)

var CreateFileUsage = NewUsage("create_file", "Create a file", []Parameter{
	NewParameter("path", "The path to the file to create", true),
	NewParameter("content", "The content of the file to create", true),
}, "<create_file><path>path/to/file.md</path><content>Hello, world!</content></create_file>")

type CreateFile struct {
	XMLName xml.Name `xml:"create_file"`
	Path    string   `xml:"path"`
	Content string   `xml:"content"`
}

func (fc CreateFile) Match(cs ChangeFileCases) (string, bool, error) { return cs.CreateFile(fc) }

func NewCreateFile(path, content string) CreateFile {
	return CreateFile{
		XMLName: xml.Name{Space: "", Local: "create_file"},
		Path:    path,
		Content: content,
	}
}
