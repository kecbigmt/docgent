package github

import (
	"context"
	"fmt"
	"path"

	"docgent-backend/internal/domain"

	"github.com/google/go-github/v68/github"
)

type FileQueryService struct {
	client     *github.Client
	owner      string
	repo       string
	baseBranch string
}

func NewFileQueryService(client *github.Client, config APIConfig) *FileQueryService {
	return &FileQueryService{
		client:     client,
		owner:      config.Owner,
		repo:       config.Repo,
		baseBranch: config.BaseBranch,
	}
}

func (s *FileQueryService) Find(name string) (domain.File, error) {
	// docsディレクトリ内のパスを構築
	filePath := path.Join("docs", name)

	// ファイルの内容を取得
	fileContent, _, _, err := s.client.Repositories.GetContents(
		context.Background(),
		s.owner,
		s.repo,
		filePath,
		&github.RepositoryContentGetOptions{
			Ref: s.baseBranch,
		},
	)
	if err != nil {
		return domain.File{}, fmt.Errorf("failed to get file contents: %w", err)
	}

	// ファイルの内容をデコード
	content, err := fileContent.GetContent()
	if err != nil {
		return domain.File{}, fmt.Errorf("failed to decode file content: %w", err)
	}

	return domain.File{
		Name:    name,
		Content: content,
	}, nil
}
