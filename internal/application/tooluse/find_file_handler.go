package tooluse

import (
	"context"
	"errors"
	"fmt"

	"docgent/internal/application/port"
	"docgent/internal/domain/tooluse"
)

// FindFileHandler は find_file ツールのハンドラーです
type FindFileHandler struct {
	ctx              context.Context
	fileQueryService port.FileQueryService
}

func NewFindFileHandler(ctx context.Context, fileQueryService port.FileQueryService) *FindFileHandler {
	return &FindFileHandler{
		ctx:              ctx,
		fileQueryService: fileQueryService,
	}
}

func (h *FindFileHandler) Handle(toolUse tooluse.FindFile) (string, bool, error) {
	file, err := h.fileQueryService.FindFile(h.ctx, toolUse.Path)
	if err != nil {
		if errors.Is(err, port.ErrFileNotFound) {
			return fmt.Sprintf("<error>File not found: %s</error>", toolUse.Path), false, nil
		}
		return "", false, err
	}
	return fmt.Sprintf("<success>\n<content>%s</content>\n</success>", file.Content), false, nil
}
