package tooluse

import "encoding/xml"

type ChangeFile struct {
	change ChangeFileUnion
}

func (fc ChangeFile) Match(cs Cases) (string, bool, error) { return cs.ChangeFile(fc) }

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
	Match(ChangeFileCases) (string, bool, error)
}

type ChangeFileCases struct {
	CreateFile func(CreateFile) (string, bool, error)
	ModifyFile func(ModifyFile) (string, bool, error)
	RenameFile func(RenameFile) (string, bool, error)
	DeleteFile func(DeleteFile) (string, bool, error)
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
