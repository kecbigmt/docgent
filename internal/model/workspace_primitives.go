package model

import "fmt"

type WorkspaceID struct {
	value string
}

func NewWorkspaceID(value string) (WorkspaceID, error) {
	return WorkspaceID{value: value}, nil
}

func (w WorkspaceID) Value() string {
	return w.value
}

func (w WorkspaceID) Equals(other WorkspaceID) bool {
	return w.value == other.value
}

type WorkspaceName struct {
	value string
}

func NewWorkspaceName(value string) (WorkspaceName, error) {
	if len(value) < 1 || len(value) > 30 {
		return WorkspaceName{}, fmt.Errorf("workspace name must be between 1 and 30 characters")
	}

	return WorkspaceName{value: value}, nil
}

func (w WorkspaceName) Value() string {
	return w.value
}

func (w WorkspaceName) Equals(other WorkspaceName) bool {
	return w.value == other.value
}
