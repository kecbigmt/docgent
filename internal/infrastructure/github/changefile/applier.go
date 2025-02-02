package changefile

import (
	"context"
	"fmt"

	"github.com/google/go-github/v68/github"

	"docgent-backend/internal/domain/autoagent/tooluse"
)

type Applier struct {
	client     *github.Client
	owner      string
	repo       string
	branchName string
}

func NewApplier(
	client *github.Client,
	owner, repo, branchName string,
) *Applier {
	return &Applier{
		client:     client,
		owner:      owner,
		repo:       repo,
		branchName: branchName,
	}
}

func (h *Applier) Apply(ctx context.Context, fc tooluse.ChangeFile) error {
	change := fc.Unwrap()
	cases := tooluse.FileChangeCases{
		CreateFile: func(c tooluse.CreateFile) error { return h.handleCreate(ctx, c) },
		ModifyFile: func(c tooluse.ModifyFile) error { return h.handleModify(ctx, c) },
		RenameFile: func(c tooluse.RenameFile) error { return h.handleRename(ctx, c) },
		DeleteFile: func(c tooluse.DeleteFile) error { return h.handleDelete(ctx, c) },
	}
	return change.Match(cases)
}

func (h *Applier) handleCreate(ctx context.Context, toolUse tooluse.CreateFile) error {
	opts := &github.RepositoryContentFileOptions{
		Message: github.Ptr(fmt.Sprintf("Create file %s", toolUse.Path)),
		Content: []byte(toolUse.Content),
		Branch:  github.Ptr(h.branchName),
	}

	_, _, err := h.client.Repositories.CreateFile(
		ctx,
		h.owner,
		h.repo,
		toolUse.Path,
		opts,
	)
	if err != nil {
		return fmt.Errorf("CreateFile failed: %w", err)
	}

	return nil
}

func (h *Applier) handleModify(ctx context.Context, toolUse tooluse.ModifyFile) error {
	// 現在のコンテンツ取得
	fileContent, _, _, err := h.client.Repositories.GetContents(
		ctx,
		h.owner,
		h.repo,
		toolUse.Path,
		&github.RepositoryContentGetOptions{
			Ref: h.branchName,
		},
	)
	if err != nil {
		if ghErr, ok := err.(*github.ErrorResponse); ok && ghErr.Response.StatusCode == 404 {
			return fmt.Errorf("%w: %s", ErrNotFound, toolUse.Path)
		}
		return fmt.Errorf("GetContents failed: %w", err)
	}

	content, err := fileContent.GetContent()
	if err != nil {
		return fmt.Errorf("GetContent failed: %w", err)
	}

	// Hunk適用
	modified, err := applyHunks(content, toolUse.Hunks)
	if err != nil {
		return fmt.Errorf("%w: %v", ErrApplyHunksFailed, err)
	}

	// 更新処理
	opts := &github.RepositoryContentFileOptions{
		Message: github.Ptr(fmt.Sprintf("Update file %s", toolUse.Path)),
		Content: []byte(modified),
		Branch:  github.Ptr(h.branchName),
		SHA:     fileContent.SHA,
	}

	_, _, err = h.client.Repositories.UpdateFile(
		ctx,
		h.owner,
		h.repo,
		toolUse.Path,
		opts,
	)
	if err != nil {
		return fmt.Errorf("UpdateFile failed: %w", err)
	}

	return nil
}

func (h *Applier) handleRename(ctx context.Context, toolUse tooluse.RenameFile) error {
	// 旧ファイル取得
	oldContent, _, _, err := h.client.Repositories.GetContents(
		ctx,
		h.owner,
		h.repo,
		toolUse.OldPath,
		&github.RepositoryContentGetOptions{Ref: h.branchName},
	)
	if err != nil {
		if ghErr, ok := err.(*github.ErrorResponse); ok && ghErr.Response.StatusCode == 404 {
			return fmt.Errorf("%w: %s", ErrNotFound, toolUse.OldPath)
		}
		return fmt.Errorf("GetContents(old) failed: %w", err)
	}

	content, err := oldContent.GetContent()
	if err != nil {
		return fmt.Errorf("GetContent(old) failed: %w", err)
	}

	// Hunk適用
	modified, err := applyHunks(content, toolUse.Hunks)
	if err != nil {
		return fmt.Errorf("%w: %v", ErrApplyHunksFailed, err)
	}

	// 新規作成
	createOpts := &github.RepositoryContentFileOptions{
		Message: github.Ptr(fmt.Sprintf("Create file %s", toolUse.NewPath)),
		Content: []byte(modified),
		Branch:  github.Ptr(h.branchName),
	}

	_, _, err = h.client.Repositories.CreateFile(
		ctx,
		h.owner,
		h.repo,
		toolUse.NewPath,
		createOpts,
	)
	if err != nil {
		return fmt.Errorf("CreateFile(new) failed: %w", err)
	}

	// 旧削除
	deleteOpts := &github.RepositoryContentFileOptions{
		Message: github.Ptr(fmt.Sprintf("Delete file %s", toolUse.OldPath)),
		Branch:  github.Ptr(h.branchName),
		SHA:     oldContent.SHA,
	}

	_, _, err = h.client.Repositories.DeleteFile(
		ctx,
		h.owner,
		h.repo,
		toolUse.OldPath,
		deleteOpts,
	)
	if err != nil {
		// 必要に応じてロールバック処理を追加
		_, _, rollbackErr := h.client.Repositories.DeleteFile(
			ctx,
			h.owner,
			h.repo,
			toolUse.NewPath,
			&github.RepositoryContentFileOptions{
				Message: github.Ptr(fmt.Sprintf("Rollback: Delete file %s", toolUse.NewPath)),
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

func (h *Applier) handleDelete(ctx context.Context, toolUse tooluse.DeleteFile) error {
	// 現在のファイル取得
	fileContent, _, _, err := h.client.Repositories.GetContents(
		ctx,
		h.owner,
		h.repo,
		toolUse.Path,
		&github.RepositoryContentGetOptions{
			Ref: h.branchName,
		},
	)
	if err != nil {
		if ghErr, ok := err.(*github.ErrorResponse); ok && ghErr.Response.StatusCode == 404 {
			return fmt.Errorf("%w: %s", ErrNotFound, toolUse.Path)
		}
		return fmt.Errorf("GetContents failed: %w", err)
	}

	// 削除処理
	opts := &github.RepositoryContentFileOptions{
		Message: github.Ptr(fmt.Sprintf("Delete file %s", toolUse.Path)),
		Branch:  github.Ptr(h.branchName),
		SHA:     fileContent.SHA,
	}

	_, _, err = h.client.Repositories.DeleteFile(
		ctx,
		h.owner,
		h.repo,
		toolUse.Path,
		opts,
	)
	if err != nil {
		return fmt.Errorf("DeleteFile failed: %w", err)
	}

	return nil
}
