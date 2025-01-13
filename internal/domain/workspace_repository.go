package domain

import "context"

type WorkspaceRepository interface {
	Save(ctx context.Context, workspace Workspace) error
	Find(ctx context.Context, id WorkspaceID) (Workspace, error)
}
