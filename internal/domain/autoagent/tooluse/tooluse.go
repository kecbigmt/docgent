package tooluse

type ToolUseUnion interface {
	Match(Cases) error
}

type Cases struct {
	ChangeFile      func(ChangeFile) error
	ReadFile        func(ReadFile) error
	AttemptComplete func(AttemptComplete) error
}
