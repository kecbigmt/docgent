package rag

import (
	"context"
	"docgent-backend/internal/domain"
)

type Corpus struct {
	client     *Client
	corpusName string
}

func NewCorpus(client *Client, corpusName string) domain.RAGCorpus {
	return &Corpus{
		client:     client,
		corpusName: corpusName,
	}
}

func (c *Corpus) Query(ctx context.Context, query string, similarityTopK int32, vectorDistanceThreshold float64) ([]domain.RAGDocument, error) {
	response, err := c.client.RetrieveContexts(ctx, c.corpusName, query, similarityTopK, vectorDistanceThreshold)
	if err != nil {
		return nil, err
	}

	documents := make([]domain.RAGDocument, len(response.Contexts))
	for i, context := range response.Contexts {
		documents[i] = domain.RAGDocument{
			Source:  context.SourceUri,
			Content: context.Text,
			Score:   context.Score,
		}
	}

	return documents, nil
}
