package domain

import (
	"context"
	"fmt"
)

type RAGService interface {
	GetCorpus(corpusName string) RAGCorpus
}

// RAGCorpus is an interface for searching for related information from existing documents.
type RAGCorpus interface {
	// Query is a method to search for related information from existing documents.
	// It returns up to 10 documents in order of relevance to the query.
	Query(ctx context.Context, query string, similarityTopK int32, vectorDistanceThreshold float64) ([]RAGDocument, error)
}

// RAGDocument is a document returned from the RAG service.
type RAGDocument struct {
	// Content is the content of the document.
	Content string
	// Source is the source of the document (e.g. file path, URL, etc.).
	Source string
	// Score is the relevance of the document to the query.
	// It ranges from 0.0 to 1.0, where 1.0 is the highest relevance.
	Score float64
}

var ErrRAGCorpusUnavailable = fmt.Errorf("rag service is unavailable")
var ErrRAGCorpusTimeout = fmt.Errorf("rag service timeout")
var ErrRAGCorpusInvalidQuery = fmt.Errorf("invalid query")
