package rag

import (
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
