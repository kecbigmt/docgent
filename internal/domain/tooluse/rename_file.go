package tooluse

import "encoding/xml"

var RenameFileUsage = NewUsage("rename_file", "Rename a file. You can also use this to move a file to another directory. Make sure to check the file content with find_file before rename_file.", []Parameter{
	NewParameter("old_path", "The exact path to the existing file to rename", true),
	NewParameter("new_path", "The new path to the file", true),
	NewParameter("hunk", "The hunk to apply to the file. The hunk is a pair of search and replace strings. Search string must be exactly matched with the content of the file. Multiple hunks can be applied to the file.", false),
}, `<rename_file>
<old_path>/path/to/file.md</old_path>
<new_path>/path/to/new_file.md</new_path>
<hunk>
<search>Hello, world!</search>
<replace>Hi, world!</replace>
</hunk>
</rename_file>`)

type RenameFile struct {
	XMLName xml.Name `xml:"rename_file"`
	OldPath string   `xml:"old_path"`
	NewPath string   `xml:"new_path"`
	Hunks   []Hunk   `xml:"hunk,omitempty"`
}

func (fc RenameFile) Match(cs ChangeFileCases) (string, bool, error) { return cs.RenameFile(fc) }

// UnmarshalXML implements xml.Unmarshaler interface
func (rf *RenameFile) UnmarshalXML(d *xml.Decoder, start xml.StartElement) error {
	type Alias RenameFile
	aux := &struct {
		*Alias
	}{
		Alias: (*Alias)(rf),
	}
	if err := d.DecodeElement(aux, &start); err != nil {
		return err
	}
	if rf.Hunks == nil {
		rf.Hunks = []Hunk{}
	}
	return nil
}

func NewRenameFile(oldPath, newPath string, hunks []Hunk) RenameFile {
	if hunks == nil {
		hunks = []Hunk{}
	}
	return RenameFile{
		XMLName: xml.Name{Space: "", Local: "rename_file"},
		OldPath: oldPath,
		NewPath: newPath,
		Hunks:   hunks,
	}
}
