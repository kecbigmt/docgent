package rag

import (
	"docgent-backend/internal/domain"
	"docgent-backend/internal/infrastructure/google/vertexai/rag/lib"
)

type Service struct {
	client *lib.Client
}

func NewService(client *lib.Client) domain.RAGService {
	return &Service{
		client: client,
	}
}

func (s *Service) GetCorpus(corpusName string) domain.RAGCorpus {
	return NewCorpus(s.client, corpusName)
}
