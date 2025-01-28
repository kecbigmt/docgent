package command

import "encoding/xml"

type DeleteFile struct {
	XMLName xml.Name `xml:"delete_file"`
	Path    string   `xml:"path"`
}

func (fc DeleteFile) Match(cs Cases) { cs.DeleteFile(fc) }

func NewDeleteFile(path string) DeleteFile {
	return DeleteFile{
		XMLName: xml.Name{Space: "", Local: "delete_file"},
		Path:    path,
	}
}
