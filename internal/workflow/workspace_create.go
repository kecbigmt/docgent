package workflow

import (
	"context"
	"docgent-backend/internal/model"
)

type CreateWorkspaceDependencies struct {
	WorkspaceRepository model.WorkspaceRepository
	Crypto              model.Crypto
}

func CreateWorkspace(ctx context.Context, deps CreateWorkspaceDependencies, dto model.WorkspaceBodyDto) (model.Workspace, error) {
	validatedBody, err := model.ParseWorkspaceBody(dto)
	if err != nil {
		return model.Workspace{}, err
	}

	createWorkspaceDeps := model.CreateWorkspaceDependencies{
		Repository: deps.WorkspaceRepository,
		Crypto:     deps.Crypto,
	}
	model.CreateWorkspace(ctx, createWorkspaceDeps, validatedBody)

	return model.Workspace{}, nil
}
