package domain

import "context"

type DocumentService struct {
	agent DocumentAgent
}

type DocumentAgent interface {
	GenerateContent(ctx context.Context, input string) (DocumentContent, error)
}

func NewDocumentService(agent DocumentAgent) *DocumentService {
	return &DocumentService{agent: agent}
}

func (s *DocumentService) GenerateContent(ctx context.Context, input string) (DocumentContent, error) {
	return s.agent.GenerateContent(ctx, input)
}
