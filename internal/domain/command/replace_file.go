package command

import "encoding/xml"

type ReplaceFile struct {
	XMLName    xml.Name `xml:"replace_file"`
	OldPath    string   `xml:"old_path"`
	NewPath    string   `xml:"new_path"`
	NewContent string   `xml:"new_content"`
}

func (fc ReplaceFile) Match(cs Cases) { cs.ReplaceFile(fc) }

func NewReplaceFile(oldPath, newPath, newContent string) ReplaceFile {
	return ReplaceFile{
		XMLName:    xml.Name{Space: "", Local: "replace_file"},
		OldPath:    oldPath,
		NewPath:    newPath,
		NewContent: newContent,
	}
}
