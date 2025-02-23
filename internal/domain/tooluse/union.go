package tooluse

type Union interface {
	Match(Cases) (string, bool, error)
}

type Cases struct {
	ChangeFile          func(ChangeFile) (string, bool, error)
	FindFile            func(FindFile) (string, bool, error)
	AttemptComplete     func(AttemptComplete) (string, bool, error)
	CreateProposal      func(CreateProposal) (string, bool, error)
	UpdateProposal      func(UpdateProposal) (string, bool, error)
	QueryRAG            func(QueryRAG) (string, bool, error)
	AddKnowledgeSources func(AddKnowledgeSources) (string, bool, error)
}
