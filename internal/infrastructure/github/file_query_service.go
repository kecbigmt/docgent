package github

import (
	"context"
	"fmt"

	"docgent-backend/internal/application/port"
	"docgent-backend/internal/domain"

	"github.com/google/go-github/v68/github"
)

type FileQueryService struct {
	client *github.Client
	owner  string
	repo   string
	branch string
}

func NewFileQueryService(client *github.Client, owner, repo, branch string) *FileQueryService {
	return &FileQueryService{
		client: client,
		owner:  owner,
		repo:   repo,
		branch: branch,
	}
}

func (s *FileQueryService) FindFile(ctx context.Context, path string) (port.File, error) {
	// ファイルの内容を取得
	fileContent, _, _, err := s.client.Repositories.GetContents(
		ctx,
		s.owner,
		s.repo,
		path,
		&github.RepositoryContentGetOptions{
			Ref: s.branch,
		},
	)
	if err != nil {
		if _, ok := err.(*github.ErrorResponse); ok && err.(*github.ErrorResponse).Response.StatusCode == 404 {
			return port.File{}, domain.ErrFileNotFound
		}
		return port.File{}, fmt.Errorf("failed to get file contents: %w", err)
	}

	// ファイルの内容をデコード
	content, err := fileContent.GetContent()
	if err != nil {
		return port.File{}, fmt.Errorf("failed to decode file content: %w", err)
	}

	return port.File{
		Path:    path,
		Content: content,
	}, nil
}

func (s *FileQueryService) GetTree(ctx context.Context, options ...port.GetTreeOption) ([]port.TreeMetadata, error) {
	treeOptions := &port.GetTreeOptions{
		Recursive: false,
		TreeSHA:   "refs/heads/" + s.branch,
	}
	for _, option := range options {
		option(treeOptions)
	}

	tree, _, err := s.client.Git.GetTree(ctx, s.owner, s.repo, treeOptions.TreeSHA, treeOptions.Recursive)
	if err != nil {
		return nil, fmt.Errorf("failed to get tree: %w", err)
	}

	treeMetadata := make([]port.TreeMetadata, 0)
	for _, entry := range tree.Entries {
		treeType := port.NodeTypeFile
		if entry.GetType() == "tree" {
			treeType = port.NodeTypeDirectory
		}
		treeMetadata = append(treeMetadata, port.TreeMetadata{
			Type: treeType,
			SHA:  entry.GetSHA(),
			Path: entry.GetPath(),
			Size: entry.GetSize(),
		})
	}
	return treeMetadata, nil
}
