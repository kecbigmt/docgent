package tooluse

import (
	"context"
	"fmt"
	"strings"

	"docgent/internal/application/port"
	"docgent/internal/domain/tooluse"
)

// QueryRAGHandler は query_rag ツールのハンドラーです
type QueryRAGHandler struct {
	ctx       context.Context
	ragCorpus port.RAGCorpus
}

func NewQueryRAGHandler(ctx context.Context, ragCorpus port.RAGCorpus) *QueryRAGHandler {
	return &QueryRAGHandler{
		ctx:       ctx,
		ragCorpus: ragCorpus,
	}
}

func (h *QueryRAGHandler) Handle(toolUse tooluse.QueryRAG) (string, bool, error) {
	if h.ragCorpus == nil {
		return "<error>RAG corpus is not set.</error>", false, nil
	}
	docs, err := h.ragCorpus.Query(h.ctx, toolUse.Query, 10, 0.7)
	if err != nil {
		return fmt.Sprintf("<error>Failed to query RAG: %s</error>", err), false, nil
	}
	if len(docs) == 0 {
		return "<success>No relevant documents found.</success>", false, nil
	}
	var result strings.Builder
	result.WriteString("<success>\n")
	for _, doc := range docs {
		result.WriteString(fmt.Sprintf("<document source=%q score=%.2f>\n%s\n</document>\n", doc.Source, doc.Score, doc.Content))
	}
	result.WriteString("</success>")
	return result.String(), false, nil
}
