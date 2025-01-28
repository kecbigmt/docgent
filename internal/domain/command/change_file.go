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
	CreateFile  func(CreateFile)
	ModifyFile  func(ModifyFile)
	ReplaceFile func(ReplaceFile)
	DeleteFile  func(DeleteFile)
}
