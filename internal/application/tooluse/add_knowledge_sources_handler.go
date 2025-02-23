package tooluse

import (
	"context"

	"docgent/internal/domain/data"
	"docgent/internal/domain/tooluse"
)

// AddKnowledgeSourcesHandler は既存のファイルに知識源情報を追加するハンドラーです
type AddKnowledgeSourcesHandler struct {
	ctx            context.Context
	fileRepository data.FileRepository
	fileChanged    *bool
}

// NewAddKnowledgeSourcesHandler は AddKnowledgeSourcesHandler の新しいインスタンスを作成します
func NewAddKnowledgeSourcesHandler(ctx context.Context, fileRepository data.FileRepository, fileChanged *bool) *AddKnowledgeSourcesHandler {
	return &AddKnowledgeSourcesHandler{
		ctx:            ctx,
		fileRepository: fileRepository,
		fileChanged:    fileChanged,
	}
}

// Handle は AddKnowledgeSources ツールの呼び出しを処理します
func (h *AddKnowledgeSourcesHandler) Handle(toolUse tooluse.AddKnowledgeSources) (string, bool, error) {
	// 既存のファイルを取得
	file, err := h.fileRepository.Get(h.ctx, toolUse.FilePath)
	if err != nil {
		return "", false, err
	}

	// 知識源情報を追加
	for _, uri := range toolUse.URIs {
		// 重複チェック
		exists := false
		for _, source := range file.KnowledgeSources {
			if source.URI == uri {
				exists = true
				break
			}
		}
		if !exists {
			file.KnowledgeSources = append(file.KnowledgeSources, data.KnowledgeSource{URI: uri})
		}
	}

	// ファイルを更新
	err = h.fileRepository.Update(h.ctx, file)
	if err != nil {
		return "", false, err
	}

	*h.fileChanged = true

	return "<success>Knowledge sources added</success>", false, nil
}
