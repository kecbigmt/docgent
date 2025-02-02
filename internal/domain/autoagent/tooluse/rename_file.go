package tooluse

import "encoding/xml"

type RenameFile struct {
	XMLName xml.Name `xml:"rename_file"`
	OldPath string   `xml:"old_path"`
	NewPath string   `xml:"new_path"`
	Hunks   []Hunk   `xml:"hunk,omitempty"`
}

func (fc RenameFile) Match(cs FileChangeCases) error { return cs.RenameFile(fc) }

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
