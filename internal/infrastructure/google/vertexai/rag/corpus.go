package rag

import (
	"context"
	"fmt"

	"docgent-backend/internal/application/port"
	"docgent-backend/internal/infrastructure/google/vertexai/rag/lib"
)

type Corpus struct {
	client   *lib.Client
	corpusId int64
}

func NewCorpus(client *lib.Client, corpusId int64) port.RAGCorpus {
	return &Corpus{
		client:   client,
		corpusId: corpusId,
	}
}

func (c *Corpus) Query(ctx context.Context, query string, similarityTopK int32, vectorDistanceThreshold float64) ([]port.RAGDocument, error) {
	response, err := c.client.RetrieveContexts(ctx, c.corpusId, query, similarityTopK, vectorDistanceThreshold)
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve contexts: %w", err)
	}

	documents := make([]port.RAGDocument, len(response.Contexts.Contexts))
	for i, context := range response.Contexts.Contexts {
		documents[i] = port.RAGDocument{
			Source:  context.SourceUri,
			Content: context.Text,
			Score:   context.Score,
		}
	}

	return documents, nil
}
