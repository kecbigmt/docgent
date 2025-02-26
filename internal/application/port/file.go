package port

import (
	"context"
	"docgent/internal/domain/data"
	"errors"
)

var (
	ErrFileNotFound = errors.New("file not found")
)

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
	FindFile(ctx context.Context, path string) (data.File, error)
	GetTree(ctx context.Context, options ...GetTreeOption) ([]TreeMetadata, error)
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
