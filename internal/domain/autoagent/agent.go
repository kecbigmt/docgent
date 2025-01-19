package autoagent

type Agent interface {
	Generate(messages []Message) (Response, error)
}
