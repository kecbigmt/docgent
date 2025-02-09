package rag

import (
	"context"
	"fmt"

	"docgent-backend/internal/application/port"
	"docgent-backend/internal/infrastructure/google/vertexai/rag/lib"
)

type Corpus struct {
	client     *lib.Client
	corpusName string
}

func NewCorpus(client *lib.Client, corpusName string) port.RAGCorpus {
	return &Corpus{
		client:     client,
		corpusName: corpusName,
	}
}

func (c *Corpus) Query(ctx context.Context, query string, similarityTopK int32, vectorDistanceThreshold float64) ([]port.RAGDocument, error) {
	response, err := c.client.RetrieveContexts(ctx, c.corpusName, query, similarityTopK, vectorDistanceThreshold)
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
