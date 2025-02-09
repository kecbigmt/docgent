package lib

import (
	"encoding/json"
	"fmt"
)

type File struct {
	Name        string     `json:"name"`
	DisplayName string     `json:"displayName"`
	Description string     `json:"description,omitempty"`
	CreateTime  string     `json:"createTime,omitempty"`
	UpdateTime  string     `json:"updateTime,omitempty"`
	FileStatus  FileStatus `json:"fileStatus,omitempty"`
}

type FileStatus struct {
	State       FileState `json:"state,omitempty"`
	ErrorStatus string    `json:"errorStatus,omitempty"`
}

type FileState int

const (
	FileStateStateUnspecified FileState = iota
	FileStateActive
	FileStateError
)

func (s FileState) String() string {
	return []string{"STATE_UNSPECIFIED", "ACTIVE", "ERROR"}[s]
}

func (s FileState) MarshalJSON() ([]byte, error) {
	return json.Marshal(s.String())
}

func (s *FileState) UnmarshalJSON(data []byte) error {
	var str string
	if err := json.Unmarshal(data, &str); err != nil {
		return err
	}
	switch str {
	case "STATE_UNSPECIFIED":
		*s = FileStateStateUnspecified
	case "ACTIVE":
		*s = FileStateActive
	case "ERROR":
		*s = FileStateError
	default:
		return fmt.Errorf("invalid file state: %s", str)
	}
	return nil
}
