package tooluse

import (
	"encoding/xml"
)

var ModifyFileUsage = NewUsage("modify_file", "Modify a existing file. Make sure to check the file content with find_file before modify_file.", []Parameter{
	NewParameter("path", "The exact path to the existing file to modify", true),
	NewParameter("hunk", "The hunk to apply to the file. The hunk is a pair of search and replace strings. Search string must be exactly matched with the content of the file. Multiple hunks can be applied to the file.", true),
}, `<modify_file>
<path>path/to/file.md</path>
<hunk>
<search>
Hello,
world!
</search>
<replace>
Hi,
world!
</replace>
</hunk>
<hunk>
<search>
Fizz
</search>
<replace>
FizzBuzz
</replace>
</hunk>
</modify_file>`)

type ModifyFile struct {
	XMLName xml.Name `xml:"modify_file"`
	Path    string   `xml:"path"`
	Hunks   []Hunk   `xml:"hunk"`
}

func (fc ModifyFile) Match(cs ChangeFileCases) (string, bool, error) { return cs.ModifyFile(fc) }

func (mf *ModifyFile) UnmarshalXML(d *xml.Decoder, start xml.StartElement) error {
	type modifyFile ModifyFile
	var v modifyFile
	if err := d.DecodeElement(&v, &start); err != nil {
		return err
	}
	if v.Hunks == nil {
		return ErrEmptyHunks
	}
	*mf = ModifyFile(v)
	return nil
}

func NewModifyFile(path string, hunks []Hunk) ModifyFile {
	return ModifyFile{
		XMLName: xml.Name{Space: "", Local: "modify_file"},
		Path:    path,
		Hunks:   hunks,
	}
}
