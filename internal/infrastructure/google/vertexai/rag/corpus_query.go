package rag

import (
	"context"
	"docgent/internal/application/port"
	"fmt"
)

func (c *Corpus) Query(ctx context.Context, query string, similarityTopK int32, vectorDistanceThreshold float64) ([]port.RAGDocument, error) {
	response, err := c.client.RetrieveContexts(ctx, c.corpusId, query, similarityTopK, vectorDistanceThreshold)
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve contexts: %w", err)
	}

	documents := make([]port.RAGDocument, len(response.Contexts.Contexts))
	for i, context := range response.Contexts.Contexts {
		documents[i] = port.RAGDocument{
			Source:  context.SourceURI,
			Content: context.Text,
			Score:   context.Score,
		}
	}

	return documents, nil
}
