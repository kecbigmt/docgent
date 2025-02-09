package port

import (
	"context"
	"docgent-backend/internal/domain/tooluse"
)

type File struct {
	Path    string
	Content string
}

type TreeMetadata struct {
	Type TreeType
	SHA  string
	Path string
	Size int
}

type TreeType int

const (
	NodeTypeFile TreeType = iota
	NodeTypeDirectory
)

type FileQueryService interface {
	FindFile(ctx context.Context, path string) (File, error)
	GetTree(ctx context.Context, options ...GetTreeOption) ([]TreeMetadata, error)
}

type FileChangeService interface {
	CreateFile(ctx context.Context, path, content string) error
	DeleteFile(ctx context.Context, path string) error
	ModifyFile(ctx context.Context, path string, hunks []tooluse.Hunk) error
	RenameFile(ctx context.Context, oldPath, newPath string, hunks []tooluse.Hunk) error
}

type GetTreeOption func(*GetTreeOptions)

type GetTreeOptions struct {
	TreeSHA   string
	Recursive bool
}

func WithGetTreeTreeSHA(treeSHA string) GetTreeOption {
	return func(o *GetTreeOptions) {
		o.TreeSHA = treeSHA
	}
}

func WithGetTreeRecursive() GetTreeOption {
	return func(o *GetTreeOptions) {
		o.Recursive = true
	}
}
