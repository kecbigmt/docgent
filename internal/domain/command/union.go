package command

type CommandUnion interface {
	Match(Cases)
}

type Cases struct {
	CreateFile  func(CreateFile)
	ModifyFile  func(ModifyFile)
	ReplaceFile func(ReplaceFile)
	DeleteFile  func(DeleteFile)
}
