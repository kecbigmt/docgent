package command

import (
	"encoding/xml"
)

type ModifyFile struct {
	XMLName xml.Name     `xml:"modify_file"`
	Path    string       `xml:"path"`
	Hunks   []ModifyHunk `xml:"hunk"`
}

func (fc ModifyFile) Match(cs Cases) { cs.ModifyFile(fc) }

func (mf *ModifyFile) UnmarshalXML(d *xml.Decoder, start xml.StartElement) error {
	type modifyFile ModifyFile
	var v modifyFile
	if err := d.DecodeElement(&v, &start); err != nil {
		return err
	}
	if v.Hunks == nil {
		return ErrEmptyModifyHunks
	}
	*mf = ModifyFile(v)
	return nil
}

func NewModifyFile(path string, hunks []ModifyHunk) ModifyFile {
	return ModifyFile{
		XMLName: xml.Name{Space: "", Local: "modify_file"},
		Path:    path,
		Hunks:   hunks,
	}
}

type ModifyHunk struct {
	Search  string `xml:"search"`
	Replace string `xml:"replace"`
}

func NewModifyHunk(search, replace string) ModifyHunk {
	return ModifyHunk{
		Search:  search,
		Replace: replace,
	}
}
