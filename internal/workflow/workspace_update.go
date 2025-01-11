package workflow

import (
	"context"
	"docgent-backend/internal/model"
)

type UpdateWorkspaceDependencies struct {
	WorkspaceRepository model.WorkspaceRepository
}

func UpdateWorkspace(ctx context.Context, deps UpdateWorkspaceDependencies, id model.WorkspaceID, dto model.WorkspaceBodyDto) error {
	validatedBody, err := model.ParseWorkspaceBody(dto)
	if err != nil {
		return err
	}

	updateWorkspaceDeps := model.UpdateWorkspaceDependencies{
		Repository: deps.WorkspaceRepository,
	}
	return model.UpdateWorkspace(ctx, updateWorkspaceDeps, id, validatedBody)
}
