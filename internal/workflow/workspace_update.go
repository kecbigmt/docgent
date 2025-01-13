package workflow

import (
	"context"

	"docgent-backend/internal/domain"
)

type UpdateWorkspaceDependencies struct {
	WorkspaceRepository domain.WorkspaceRepository
}

func UpdateWorkspace(ctx context.Context, deps UpdateWorkspaceDependencies, id domain.WorkspaceID, dto domain.WorkspaceBodyDto) error {
	validatedBody, err := domain.ParseWorkspaceBody(dto)
	if err != nil {
		return err
	}

	updateWorkspaceDeps := domain.UpdateWorkspaceDependencies{
		Repository: deps.WorkspaceRepository,
	}
	return domain.UpdateWorkspace(ctx, updateWorkspaceDeps, id, validatedBody)
}
