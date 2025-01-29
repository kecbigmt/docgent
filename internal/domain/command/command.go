package command

type CommandUnion interface {
	Match(Cases) error
}

type Cases struct {
	ChangeFile func(ChangeFile) error
	ReadFile   func(ReadFile) error
}
