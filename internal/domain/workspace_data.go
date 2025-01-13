package domain

type Workspace struct {
	ID   WorkspaceID
	Body WorkspaceBody
}

type WorkspaceBody struct {
	name WorkspaceName
}

func NewWorkspaceBody(name WorkspaceName) (WorkspaceBody, error) {
	return WorkspaceBody{name: name}, nil
}

type WorkspaceBodyDto struct {
	Name string
}
