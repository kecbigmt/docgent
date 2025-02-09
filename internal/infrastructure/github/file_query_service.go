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
