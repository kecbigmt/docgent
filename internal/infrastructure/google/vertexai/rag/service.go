package rag

import (
	"docgent-backend/internal/domain"
)

type Service struct {
	client *Client
}

func NewService(client *Client) domain.RAGService {
	return &Service{
		client: client,
	}
}

func (s *Service) GetCorpus(corpusName string) domain.RAGCorpus {
	return NewCorpus(s.client, corpusName)
}
