package workflow

import (
	"context"

	"docgent-backend/internal/domain"
)

type CreateWorkspaceDependencies struct {
	WorkspaceRepository domain.WorkspaceRepository
	Crypto              domain.Crypto
}

func CreateWorkspace(ctx context.Context, deps CreateWorkspaceDependencies, dto domain.WorkspaceBodyDto) (domain.Workspace, error) {
	validatedBody, err := domain.ParseWorkspaceBody(dto)
	if err != nil {
		return domain.Workspace{}, err
	}

	createWorkspaceDeps := domain.CreateWorkspaceDependencies{
		Repository: deps.WorkspaceRepository,
		Crypto:     deps.Crypto,
	}
	domain.CreateWorkspace(ctx, createWorkspaceDeps, validatedBody)

	return domain.Workspace{}, nil
}
