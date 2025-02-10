package rag

import (
	"docgent/internal/application/port"
	"docgent/internal/infrastructure/google/vertexai/rag/lib"
)

type Service struct {
	client *lib.Client
}

func NewService(client *lib.Client) port.RAGService {
	return &Service{
		client: client,
	}
}

func (s *Service) GetCorpus(corpusId int64) port.RAGCorpus {
	return NewCorpus(s.client, corpusId)
}
