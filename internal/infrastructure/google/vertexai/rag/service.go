package rag

import (
	"docgent-backend/internal/application/port"
	"docgent-backend/internal/infrastructure/google/vertexai/rag/lib"
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
