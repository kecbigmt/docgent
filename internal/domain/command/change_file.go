package command

import "encoding/xml"

type ChangeFile struct {
	change ChangeFileUnion
}

func (fc ChangeFile) Match(cs Cases) error { return cs.ChangeFile(fc) }

func (fc ChangeFile) Unwrap() ChangeFileUnion {
	return fc.change
}

// MarshalXML implements xml.Marshaler interface
func (fc ChangeFile) MarshalXML(e *xml.Encoder, start xml.StartElement) error {
	return e.Encode(fc.change)
}

func NewChangeFile(changeFile ChangeFileUnion) ChangeFile {
	return ChangeFile{change: changeFile}
}

type ChangeFileUnion interface {
	Match(FileChangeCases) error
}

type FileChangeCases struct {
	CreateFile func(CreateFile) error
	ModifyFile func(ModifyFile) error
	RenameFile func(RenameFile) error
	DeleteFile func(DeleteFile) error
}

type Hunk struct {
	Search  string `xml:"search"`
	Replace string `xml:"replace"`
}

func NewHunk(search, replace string) Hunk {
	return Hunk{
		Search:  search,
		Replace: replace,
	}
}
