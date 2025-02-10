package github

import (
	"context"
	"fmt"

	"github.com/google/go-github/v68/github"

	"docgent/internal/application/port"
	"docgent/internal/domain/tooluse"
)

type FileChangeService struct {
	client     *github.Client
	owner      string
	repo       string
	branchName string
}

func NewFileChangeService(
	client *github.Client,
	owner, repo, branchName string,
) port.FileChangeService {
	return &FileChangeService{
		client:     client,
		owner:      owner,
		repo:       repo,
		branchName: branchName,
	}
}

func (h *FileChangeService) CreateFile(ctx context.Context, path, content string) error {
	opts := &github.RepositoryContentFileOptions{
		Message: github.Ptr(fmt.Sprintf("Create file %s", path)),
		Content: []byte(content),
		Branch:  github.Ptr(h.branchName),
	}

	_, _, err := h.client.Repositories.CreateFile(
		ctx,
		h.owner,
		h.repo,
		path,
		opts,
	)
	if err != nil {
		return fmt.Errorf("CreateFile failed: %w", err)
	}

	return nil
}

func (h *FileChangeService) ModifyFile(ctx context.Context, path string, hunks []tooluse.Hunk) error {
	// 現在のコンテンツ取得
	fileContent, _, _, err := h.client.Repositories.GetContents(
		ctx,
		h.owner,
		h.repo,
		path,
		&github.RepositoryContentGetOptions{
			Ref: h.branchName,
		},
	)
	if err != nil {
		if ghErr, ok := err.(*github.ErrorResponse); ok && ghErr.Response.StatusCode == 404 {
			return fmt.Errorf("%w: %s", ErrNotFound, path)
		}
		return fmt.Errorf("GetContents failed: %w", err)
	}

	content, err := fileContent.GetContent()
	if err != nil {
		return fmt.Errorf("GetContent failed: %w", err)
	}

	// Hunk適用
	modified, err := applyHunks(content, hunks)
	if err != nil {
		return fmt.Errorf("%w: %v", ErrApplyHunksFailed, err)
	}

	// 更新処理
	opts := &github.RepositoryContentFileOptions{
		Message: github.Ptr(fmt.Sprintf("Update file %s", path)),
		Content: []byte(modified),
		Branch:  github.Ptr(h.branchName),
		SHA:     fileContent.SHA,
	}

	_, _, err = h.client.Repositories.UpdateFile(
		ctx,
		h.owner,
		h.repo,
		path,
		opts,
	)
	if err != nil {
		return fmt.Errorf("UpdateFile failed: %w", err)
	}

	return nil
}

func (h *FileChangeService) RenameFile(ctx context.Context, oldPath, newPath string, hunks []tooluse.Hunk) error {
	// 旧ファイル取得
	oldContent, _, _, err := h.client.Repositories.GetContents(
		ctx,
		h.owner,
		h.repo,
		oldPath,
		&github.RepositoryContentGetOptions{Ref: h.branchName},
	)
	if err != nil {
		if ghErr, ok := err.(*github.ErrorResponse); ok && ghErr.Response.StatusCode == 404 {
			return fmt.Errorf("%w: %s", ErrNotFound, oldPath)
		}
		return fmt.Errorf("GetContents(old) failed: %w", err)
	}

	content, err := oldContent.GetContent()
	if err != nil {
		return fmt.Errorf("GetContent(old) failed: %w", err)
	}

	// Hunk適用
	modified, err := applyHunks(content, hunks)
	if err != nil {
		return fmt.Errorf("%w: %v", ErrApplyHunksFailed, err)
	}

	// 新規作成
	createOpts := &github.RepositoryContentFileOptions{
		Message: github.Ptr(fmt.Sprintf("Create file %s", newPath)),
		Content: []byte(modified),
		Branch:  github.Ptr(h.branchName),
	}

	_, _, err = h.client.Repositories.CreateFile(
		ctx,
		h.owner,
		h.repo,
		newPath,
		createOpts,
	)
	if err != nil {
		return fmt.Errorf("CreateFile(new) failed: %w", err)
	}

	// 旧削除
	deleteOpts := &github.RepositoryContentFileOptions{
		Message: github.Ptr(fmt.Sprintf("Delete file %s", oldPath)),
		Branch:  github.Ptr(h.branchName),
		SHA:     oldContent.SHA,
	}

	_, _, err = h.client.Repositories.DeleteFile(
		ctx,
		h.owner,
		h.repo,
		oldPath,
		deleteOpts,
	)
	if err != nil {
		// 必要に応じてロールバック処理を追加
		_, _, rollbackErr := h.client.Repositories.DeleteFile(
			ctx,
			h.owner,
			h.repo,
			newPath,
			&github.RepositoryContentFileOptions{
				Message: github.Ptr(fmt.Sprintf("Rollback: Delete file %s", newPath)),
				Branch:  github.Ptr(h.branchName),
			},
		)
		if rollbackErr != nil {
			return fmt.Errorf("DeleteFile(old) failed and rollback failed: %w, rollback error: %v", err, rollbackErr)
		}
		return fmt.Errorf("DeleteFile(old) failed: %w", err)
	}

	return nil
}

func (h *FileChangeService) DeleteFile(ctx context.Context, path string) error {
	// 現在のファイル取得
	fileContent, _, _, err := h.client.Repositories.GetContents(
		ctx,
		h.owner,
		h.repo,
		path,
		&github.RepositoryContentGetOptions{
			Ref: h.branchName,
		},
	)
	if err != nil {
		if ghErr, ok := err.(*github.ErrorResponse); ok && ghErr.Response.StatusCode == 404 {
			return fmt.Errorf("%w: %s", ErrNotFound, path)
		}
		return fmt.Errorf("GetContents failed: %w", err)
	}

	// 削除処理
	opts := &github.RepositoryContentFileOptions{
		Message: github.Ptr(fmt.Sprintf("Delete file %s", path)),
		Branch:  github.Ptr(h.branchName),
		SHA:     fileContent.SHA,
	}

	_, _, err = h.client.Repositories.DeleteFile(
		ctx,
		h.owner,
		h.repo,
		path,
		opts,
	)
	if err != nil {
		return fmt.Errorf("DeleteFile failed: %w", err)
	}

	return nil
}
