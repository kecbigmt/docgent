package workflow

import (
	"context"

	"docgent-backend/internal/domain"
)

type DocumentGenerateWorkflow struct {
	documentAgent      domain.DocumentAgent
	documentRepository domain.DocumentRepository
}

func NewDocumentGenerateWorkflow(
	documentAgent domain.DocumentAgent,
	documentRepository domain.DocumentRepository,
) *DocumentGenerateWorkflow {
	return &DocumentGenerateWorkflow{
		documentAgent:      documentAgent,
		documentRepository: documentRepository,
	}
}

func (w *DocumentGenerateWorkflow) Execute(ctx context.Context, text string) (domain.Document, error) {
	documentService := domain.NewDocumentService(w.documentAgent, w.documentRepository)

	doc, err := documentService.Generate(ctx, text)
	if err != nil {
		return domain.Document{}, err
	}

	return documentService.Create(doc)
}
