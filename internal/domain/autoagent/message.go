package autoagent

type Message struct {
	Role    Role
	Content string
}

func NewMessage(role Role, content string) Message {
	return Message{Role: role, Content: content}
}

type Role int

const (
	UserRole Role = iota
	AssistantRole
)

func (r Role) String() string {
	switch r {
	case UserRole:
		return "user"
	case AssistantRole:
		return "assistant"
	default:
		return "unknown"
	}
}
