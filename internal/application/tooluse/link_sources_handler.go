package tooluse

import (
	"context"

	"docgent/internal/domain/data"
	"docgent/internal/domain/tooluse"
)

// LinkSourcesHandler は既存のファイルに知識源情報を追加するハンドラーです
type LinkSourcesHandler struct {
	ctx            context.Context
	fileRepository data.FileRepository
	fileChanged    *bool
}

// NewLinkSourcesHandler は LinkSourcesHandler の新しいインスタンスを作成します
func NewLinkSourcesHandler(ctx context.Context, fileRepository data.FileRepository, fileChanged *bool) *LinkSourcesHandler {
	return &LinkSourcesHandler{
		ctx:            ctx,
		fileRepository: fileRepository,
		fileChanged:    fileChanged,
	}
}

// Handle は AddKnowledgeSources ツールの呼び出しを処理します
func (h *LinkSourcesHandler) Handle(toolUse tooluse.LinkSources) (string, bool, error) {
	// 既存のファイルを取得
	file, err := h.fileRepository.Get(h.ctx, toolUse.FilePath)
	if err != nil {
		return "", false, err
	}

	// 知識源情報を追加
	for _, rawURI := range toolUse.URIs {
		// バリデーション
		uri, err := data.NewURI(rawURI)
		if err != nil {
			return "", false, err
		}

		// 重複チェック
		exists := false
		for _, existingURI := range file.SourceURIs {
			if existingURI.Equal(uri) {
				exists = true
				break
			}
		}
		if !exists {
			file.SourceURIs = append(file.SourceURIs, uri)
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
