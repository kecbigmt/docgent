package tooluse

import (
	"context"
	"strings"

	"docgent/internal/domain/data"
	"docgent/internal/domain/tooluse"
)

// FileChangeHandler は create_file, modify_file, rename_file, delete_file ツールのハンドラーです
type FileChangeHandler struct {
	ctx            context.Context
	fileRepository data.FileRepository
	fileChanged    *bool
}

func NewFileChangeHandler(ctx context.Context, fileRepository data.FileRepository, fileChanged *bool) *FileChangeHandler {
	return &FileChangeHandler{
		ctx:            ctx,
		fileRepository: fileRepository,
		fileChanged:    fileChanged,
	}
}

func (h *FileChangeHandler) Handle(toolUse tooluse.ChangeFile) (string, bool, error) {
	change := toolUse.Unwrap()
	cases := tooluse.ChangeFileCases{
		CreateFile: h.handleCreateFile,
		ModifyFile: h.handleModifyFile,
		RenameFile: h.handleRenameFile,
		DeleteFile: h.handleDeleteFile,
	}
	return change.Match(cases)
}

func (h *FileChangeHandler) handleCreateFile(c tooluse.CreateFile) (string, bool, error) {
	file := &data.File{
		Path:             c.Path,
		Content:          c.Content,
		KnowledgeSources: make([]data.KnowledgeSource, len(c.KnowledgeSourceURIs)),
	}
	for i, uri := range c.KnowledgeSourceURIs {
		file.KnowledgeSources[i] = data.KnowledgeSource{URI: uri}
	}

	err := h.fileRepository.Create(h.ctx, file)
	if err != nil {
		return "", false, err
	}

	*h.fileChanged = true

	return "<success>File created</success>", false, nil
}

func (h *FileChangeHandler) handleModifyFile(c tooluse.ModifyFile) (string, bool, error) {
	// 既存のファイルを取得
	file, err := h.fileRepository.Get(h.ctx, c.Path)
	if err != nil {
		return "", false, err
	}

	// ハンクを適用
	content := file.Content
	for _, hunk := range c.Hunks {
		content = strings.Replace(content, hunk.Search, hunk.Replace, 1)
	}
	file.Content = content

	// ファイルを更新
	err = h.fileRepository.Update(h.ctx, file)
	if err != nil {
		return "", false, err
	}

	*h.fileChanged = true

	return "<success>File modified</success>", false, nil
}

func (h *FileChangeHandler) handleRenameFile(c tooluse.RenameFile) (string, bool, error) {
	// 既存のファイルを取得
	file, err := h.fileRepository.Get(h.ctx, c.OldPath)
	if err != nil {
		return "", false, err
	}

	// ハンクを適用
	content := file.Content
	for _, hunk := range c.Hunks {
		content = strings.Replace(content, hunk.Search, hunk.Replace, 1)
	}

	// 新しいファイルを作成
	newFile := &data.File{
		Path:             c.NewPath,
		Content:          content,
		KnowledgeSources: file.KnowledgeSources,
	}
	err = h.fileRepository.Create(h.ctx, newFile)
	if err != nil {
		return "", false, err
	}

	// 古いファイルを削除
	err = h.fileRepository.Delete(h.ctx, c.OldPath)
	if err != nil {
		return "", false, err
	}

	*h.fileChanged = true

	return "<success>File renamed</success>", false, nil
}

func (h *FileChangeHandler) handleDeleteFile(c tooluse.DeleteFile) (string, bool, error) {
	err := h.fileRepository.Delete(h.ctx, c.Path)
	if err != nil {
		return "", false, err
	}

	*h.fileChanged = true

	return "<success>File deleted</success>", false, nil
}
