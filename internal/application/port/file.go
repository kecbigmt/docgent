package port

import (
	"context"
	"errors"

	"docgent/internal/domain/data"
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
	// GetURI returns a URI for the given file path, typically a GitHub permalink
	GetURI(ctx context.Context, path string) (*data.URI, error)
	// GetFilePath returns the file path for the given URI
	GetFilePath(uri *data.URI) (string, error)
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
