package tooluse

import "encoding/xml"

var DeleteFileUsage = NewUsage("delete_file", "Delete a file", []Parameter{
	NewParameter("path", "The exact path to the existing file to delete", true),
}, "<delete_file><path>path/to/file.md</path></delete_file>")

type DeleteFile struct {
	XMLName xml.Name `xml:"delete_file"`
	Path    string   `xml:"path"`
}

func (fc DeleteFile) Match(cs ChangeFileCases) (string, bool, error) { return cs.DeleteFile(fc) }

func NewDeleteFile(path string) DeleteFile {
	return DeleteFile{
		XMLName: xml.Name{Space: "", Local: "delete_file"},
		Path:    path,
	}
}
