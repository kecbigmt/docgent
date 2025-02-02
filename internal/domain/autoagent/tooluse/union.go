package tooluse

type Union interface {
	Match(Cases) (string, bool, error)
}

type Cases struct {
	ChangeFile      func(ChangeFile) (string, bool, error)
	ReadFile        func(ReadFile) (string, bool, error)
	AttemptComplete func(AttemptComplete) (string, bool, error)
}
