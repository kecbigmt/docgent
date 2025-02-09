package lib

import (
	"encoding/json"
	"fmt"
)

type Corpus struct {
	Name         string       `json:"name"`
	DisplayName  string       `json:"displayName"`
	Description  string       `json:"description,omitempty"`
	CreateTime   string       `json:"createTime,omitempty"`
	UpdateTime   string       `json:"updateTime,omitempty"`
	CorpusStatus CorpusStatus `json:"corpusStatus,omitempty"`
}

type CorpusStatus struct {
	State       CorpusState `json:"state,omitempty"`
	ErrorStatus string      `json:"errorStatus,omitempty"`
}

type CorpusState int

const (
	CorpusStateUnknown CorpusState = iota
	CorpusStateInitialized
	CorpusStateActive
	CorpusStateError
)

func (s CorpusState) MarshalJSON() ([]byte, error) {
	return json.Marshal(s.String())
}

func (s CorpusState) String() string {
	return []string{"UNKNOWN", "INITIALIZED", "ACTIVE", "ERROR"}[s]
}

func (s *CorpusState) UnmarshalJSON(data []byte) error {
	var str string
	if err := json.Unmarshal(data, &str); err != nil {
		return err
	}
	switch str {
	case "UNKNOWN":
		*s = CorpusStateUnknown
	case "INITIALIZED":
		*s = CorpusStateInitialized
	case "ACTIVE":
		*s = CorpusStateActive
	case "ERROR":
		*s = CorpusStateError
	default:
		return fmt.Errorf("invalid corpus state: %s", str)
	}
	return nil
}
