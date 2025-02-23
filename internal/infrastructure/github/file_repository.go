package github

import (
	"context"
	"fmt"
	"net/http"

	"docgent/internal/domain/data"
	"docgent/internal/infrastructure/yaml"

	"github.com/google/go-github/v68/github"
)

// FileRepository はGitHub上でのファイル操作を実装
type FileRepository struct {
	client *github.Client
	owner  string
	repo   string
	branch string
}

func NewFileRepository(client *github.Client, owner, repo, branch string) *FileRepository {
	return &FileRepository{
		client: client,
		owner:  owner,
		repo:   repo,
		branch: branch,
	}
}

// Create はファイルを作成します
func (r *FileRepository) Create(ctx context.Context, file *data.File) error {
	// 知識ソースのバリデーション
	for _, ks := range file.KnowledgeSources {
		if ks.URI == "" {
			return data.ErrInvalidKnowledgeSource
		}
	}

	// ファイルの存在確認
	_, _, resp, err := r.client.Repositories.GetContents(ctx, r.owner, r.repo, file.Path, &github.RepositoryContentGetOptions{
		Ref: r.branch,
	})
	if err == nil {
		return data.ErrFileAlreadyExists
	}
	if resp.StatusCode != http.StatusNotFound {
		return fmt.Errorf("%w: %d %s %v", data.ErrFailedToAccessFile, resp.StatusCode, resp.Status, err)
	}

	// フロントマターの生成
	frontmatter, err := yaml.GenerateFrontmatter(file.KnowledgeSources)
	if err != nil {
		return fmt.Errorf("failed to generate frontmatter: %w", err)
	}

	// フロントマターと本文の結合
	content := yaml.CombineContentAndFrontmatter(frontmatter, file.Content)

	// ファイルの作成
	_, _, err = r.client.Repositories.CreateFile(ctx, r.owner, r.repo, file.Path, &github.RepositoryContentFileOptions{
		Message: github.Ptr(fmt.Sprintf("Create file %s", file.Path)),
		Content: []byte(content),
		Branch:  github.Ptr(r.branch),
	})
	if err != nil {
		return fmt.Errorf("failed to create file: %w", err)
	}

	return nil
}

func (r *FileRepository) Update(ctx context.Context, file *data.File) error {
	// 知識ソースのバリデーション
	for _, ks := range file.KnowledgeSources {
		if ks.URI == "" {
			return data.ErrInvalidKnowledgeSource
		}
	}

	// YAMLフロントマターを生成
	frontmatter, err := yaml.GenerateFrontmatter(file.KnowledgeSources)
	if err != nil {
		return fmt.Errorf("%w: %s", data.ErrInvalidKnowledgeSource, err.Error())
	}

	// 現在のファイルを取得
	fileContent, _, _, err := r.client.Repositories.GetContents(
		ctx,
		r.owner,
		r.repo,
		file.Path,
		&github.RepositoryContentGetOptions{
			Ref: r.branch,
		},
	)
	if err != nil {
		if ghErr, ok := err.(*github.ErrorResponse); ok && ghErr.Response.StatusCode == 404 {
			return fmt.Errorf("%w: %s", data.ErrFileNotFound, file.Path)
		}
		return fmt.Errorf("%w: %s", data.ErrFailedToAccessFile, err.Error())
	}

	// フロントマターとコンテンツを結合
	content := yaml.CombineContentAndFrontmatter(frontmatter, file.Content)

	// GitHubのファイルを更新
	opts := &github.RepositoryContentFileOptions{
		Message: github.Ptr(fmt.Sprintf("Update file %s", file.Path)),
		Content: []byte(content),
		Branch:  github.Ptr(r.branch),
		SHA:     fileContent.SHA,
	}

	_, _, err = r.client.Repositories.UpdateFile(
		ctx,
		r.owner,
		r.repo,
		file.Path,
		opts,
	)
	if err != nil {
		return fmt.Errorf("%w: %s", data.ErrFailedToAccessFile, err.Error())
	}

	return nil
}

func (r *FileRepository) Get(ctx context.Context, path string) (*data.File, error) {
	// GitHubからファイルを取得
	fileContent, _, _, err := r.client.Repositories.GetContents(
		ctx,
		r.owner,
		r.repo,
		path,
		&github.RepositoryContentGetOptions{
			Ref: r.branch,
		},
	)
	if err != nil {
		if ghErr, ok := err.(*github.ErrorResponse); ok && ghErr.Response.StatusCode == 404 {
			return nil, fmt.Errorf("%w: %s", data.ErrFileNotFound, path)
		}
		return nil, fmt.Errorf("%w: %s", data.ErrFailedToAccessFile, err.Error())
	}

	content, err := fileContent.GetContent()
	if err != nil {
		return nil, fmt.Errorf("%w: %s", data.ErrFailedToAccessFile, err.Error())
	}

	// フロントマターとコンテンツを分離
	frontmatter, body, err := yaml.SplitContentAndFrontmatter(content)
	if err != nil {
		return nil, fmt.Errorf("%w: %s", data.ErrInvalidFrontmatter, err.Error())
	}

	// フロントマーターをパース
	var sources []data.KnowledgeSource
	if frontmatter != "" {
		sources, err = yaml.ParseFrontmatter(frontmatter)
		if err != nil {
			return nil, fmt.Errorf("%w: %s", data.ErrInvalidFrontmatter, err.Error())
		}
	}

	return &data.File{
		Path:             path,
		Content:          body,
		KnowledgeSources: sources,
	}, nil
}

func (r *FileRepository) Delete(ctx context.Context, path string) error {
	// 現在のファイルを取得
	fileContent, _, _, err := r.client.Repositories.GetContents(
		ctx,
		r.owner,
		r.repo,
		path,
		&github.RepositoryContentGetOptions{
			Ref: r.branch,
		},
	)
	if err != nil {
		if ghErr, ok := err.(*github.ErrorResponse); ok && ghErr.Response.StatusCode == 404 {
			return fmt.Errorf("%w: %s", data.ErrFileNotFound, path)
		}
		return fmt.Errorf("%w: %s", data.ErrFailedToAccessFile, err.Error())
	}

	// GitHubのファイルを削除
	opts := &github.RepositoryContentFileOptions{
		Message: github.Ptr(fmt.Sprintf("Delete file %s", path)),
		Branch:  github.Ptr(r.branch),
		SHA:     fileContent.SHA,
	}

	_, _, err = r.client.Repositories.DeleteFile(
		ctx,
		r.owner,
		r.repo,
		path,
		opts,
	)
	if err != nil {
		return fmt.Errorf("%w: %s", data.ErrFailedToAccessFile, err.Error())
	}

	return nil
}
