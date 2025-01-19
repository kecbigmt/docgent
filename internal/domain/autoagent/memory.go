package autoagent

import "fmt"

type MemoryMap map[string]string

func (m MemoryMap) ToXMLString() string {
	str := "<memory_map>"
	for key, value := range m {
		str += fmt.Sprintf("<item><key>%s</key><value>%s</value></item>", key, value)
	}
	str += "</memory_map>"
	return str
}

type LongTermMemoryService interface {
	Save(key, value string) (MemoryMap, error)
	Delete(key string) (MemoryMap, error)
	RetriveAll() (MemoryMap, error)
}
