package model

import (
	"context"

	"docgent-backend/internal/model/infrastructure"
)

func ParseWorkspaceBody(dto WorkspaceBodyDto) (WorkspaceBody, error) {
	name, err := NewWorkspaceName(dto.Name)
	if err != nil {
		return WorkspaceBody{}, err
	}
	body, err := NewWorkspaceBody(name)
	if err != nil {
		return WorkspaceBody{}, err
	}

	return body, nil
}

type CreateWorkspaceDependencies struct {
	Repository WorkspaceRepository
	Crypto     infrastructure.Crypto
}

func CreateWorkspace(ctx context.Context, deps CreateWorkspaceDependencies, body WorkspaceBody) (Workspace, error) {
	rawId, err := deps.Crypto.GenerateID()
	if err != nil {
		return Workspace{}, err
	}

	id, err := NewWorkspaceID(rawId)
	if err != nil {
		return Workspace{}, err
	}

	workspace := Workspace{id, body}
	err = deps.Repository.Save(ctx, workspace)
	if err != nil {
		return Workspace{}, err
	}

	return workspace, nil
}

type UpdateWorkspaceDependencies struct {
	Repository WorkspaceRepository
}

func UpdateWorkspace(ctx context.Context, deps UpdateWorkspaceDependencies, id WorkspaceID, body WorkspaceBody) error {
	workspace, err := deps.Repository.Find(ctx, id)
	if err != nil {
		return err
	}

	workspace.Body = body
	return deps.Repository.Save(ctx, workspace)
}
