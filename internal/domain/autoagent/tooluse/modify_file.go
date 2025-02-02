package tooluse

import (
	"encoding/xml"
)

type ModifyFile struct {
	XMLName xml.Name `xml:"modify_file"`
	Path    string   `xml:"path"`
	Hunks   []Hunk   `xml:"hunk"`
}

func (fc ModifyFile) Match(cs FileChangeCases) (string, bool, error) { return cs.ModifyFile(fc) }

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
