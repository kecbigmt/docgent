package rag

import (
	"docgent/internal/application/port"
	"docgent/internal/infrastructure/google/vertexai/rag/lib"
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
