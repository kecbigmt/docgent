package port

import (
	"context"
	"io"

	"docgent/internal/domain/data"
)

type RAGService interface {
	GetCorpus(corpusId int64) RAGCorpus
}

// RAGCorpus is an interface for searching for related information from existing documents.
type RAGCorpus interface {
	// Query is a method to search for related information from existing documents.
	// It returns up to 10 documents in order of relevance to the query.
	Query(ctx context.Context, query string, similarityTopK int32, vectorDistanceThreshold float64) ([]RAGDocument, error)

	UploadFile(ctx context.Context, file io.Reader, uri *data.URI, options ...RAGCorpusUploadFileOption) error

	ListFiles(ctx context.Context) ([]RAGFile, error)

	DeleteFile(ctx context.Context, fileID int64) error
}

type RAGCorpusUploadFileOption func(*RAGCorpusUploadFileOptions)

type RAGCorpusUploadFileOptions struct {
	Description    string
	ChunkingConfig ChunkingConfig
}

func WithRagFileDescription(description string) RAGCorpusUploadFileOption {
	return func(o *RAGCorpusUploadFileOptions) {
		o.Description = description
	}
}

func WithRagFileChunkingConfig(chunkSize int, chunkOverlap int) RAGCorpusUploadFileOption {
	return func(o *RAGCorpusUploadFileOptions) {
		o.ChunkingConfig = ChunkingConfig{
			ChunkSize:    chunkSize,
			ChunkOverlap: chunkOverlap,
		}
	}
}

type ChunkingConfig struct {
	ChunkSize    int
	ChunkOverlap int
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

type RAGFile struct {
	ID          int64
	URI         *data.URI
	Description string
}
