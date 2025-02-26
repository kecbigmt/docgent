package tooluse

import (
	"context"
	"docgent/internal/application/port"
	"docgent/internal/domain/data"
	"docgent/internal/domain/tooluse"
	"fmt"
)

type FindSourceHandler struct {
	ctx                     context.Context
	sourceRepositoryManager *port.SourceRepositoryManager
}

func NewFindSourceHandler(ctx context.Context, sourceRepositoryManager *port.SourceRepositoryManager) *FindSourceHandler {
	return &FindSourceHandler{
		ctx:                     ctx,
		sourceRepositoryManager: sourceRepositoryManager,
	}
}

func (h *FindSourceHandler) Handle(toolUse tooluse.FindSource) (string, bool, error) {
	uri, err := data.NewURI(toolUse.URI)
	if err != nil {
		return fmt.Sprintf("<error>Invalid URI: %s</error>", toolUse.URI), false, nil
	}

	source, err := h.sourceRepositoryManager.Find(h.ctx, uri)
	if err != nil {
		return fmt.Sprintf("<error>Source not found: %s</error>", uri), false, nil
	}

	return fmt.Sprintf("<success>\n<content>%s</content>\n</success>", source.Content()), false, nil
}
