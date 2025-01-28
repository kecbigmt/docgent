package command

import (
	"encoding/xml"
)

type CreateFile struct {
	XMLName xml.Name `xml:"create_file"`
	Path    string   `xml:"path"`
	Content string   `xml:"content"`
}

func (fc CreateFile) Match(cs Cases) { cs.CreateFile(fc) }

func NewCreateFile(path, content string) CreateFile {
	return CreateFile{
		XMLName: xml.Name{Space: "", Local: "create_file"},
		Path:    path,
		Content: content,
	}
}
