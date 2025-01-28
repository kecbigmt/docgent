package command

import "encoding/xml"

type ChangeFile struct {
	change ChangeFileUnion
}

func (fc ChangeFile) Match(cs Cases) { cs.ChangeFile(fc) }

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
	Match(FileChangeCases)
}

type FileChangeCases struct {
	CreateFile func(CreateFile)
	ModifyFile func(ModifyFile)
	RenameFile func(RenameFile)
	DeleteFile func(DeleteFile)
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
