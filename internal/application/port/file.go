package port

import (
	"context"
	"docgent-backend/internal/domain/tooluse"
)

type File struct {
	Path    string
	Content string
}

type FileQueryService interface {
	FindFile(ctx context.Context, path string) (File, error)
}

type FileChangeService interface {
	CreateFile(ctx context.Context, path, content string) error
	DeleteFile(ctx context.Context, path string) error
	ModifyFile(ctx context.Context, path string, hunks []tooluse.Hunk) error
	RenameFile(ctx context.Context, oldPath, newPath string, hunks []tooluse.Hunk) error
}
