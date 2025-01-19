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
	SystemRole Role = iota
	UserRole
	AgentRole
)

func (r Role) String() string {
	switch r {
	case SystemRole:
		return "system"
	case UserRole:
		return "user"
	case AgentRole:
		return "agent"
	default:
		return "unknown"
	}
}
