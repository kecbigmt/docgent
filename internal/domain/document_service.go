package domain

import "context"

type DocumentService struct {
	agent      DocumentAgent
	repository DocumentRepository
}

type DocumentRepository interface {
	Save(document Document) (Document, error)
}

type DocumentAgent interface {
	Generate(ctx context.Context, input string) (Document, error)
}

func NewDocumentService(agent DocumentAgent, repository DocumentRepository) *DocumentService {
	return &DocumentService{agent: agent, repository: repository}
}

func (s *DocumentService) Create(document Document) (Document, error) {
	return s.repository.Save(Document(document))
}

func (s *DocumentService) Generate(ctx context.Context, input string) (Document, error) {
	return s.agent.Generate(ctx, input)
}
