package command

type CommandUnion interface {
	Match(Cases)
}

type Cases struct {
	ChangeFile func(ChangeFile)
	ReadFile   func(ReadFile)
}
