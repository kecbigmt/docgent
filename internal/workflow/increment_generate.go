package workflow

import (
	"context"
	"docgent-backend/internal/domain"
)

type IncrementGenerateWorkflow struct {
	documentAgent       domain.DocumentAgent
	incrementRepository domain.IncrementRepository
}

func NewIncrementGenerateWorkflow(
	documentAgent domain.DocumentAgent,
	incrementRepository domain.IncrementRepository,
) *IncrementGenerateWorkflow {
	return &IncrementGenerateWorkflow{
		documentAgent:       documentAgent,
		incrementRepository: incrementRepository,
	}
}

func (w *IncrementGenerateWorkflow) Execute(
	ctx context.Context,
	text string,
	previousIncrementHandle domain.IncrementHandle,
) (domain.Increment, error) {
	documentService := domain.NewDocumentService(w.documentAgent)
	incrementService := domain.NewIncrementService(w.incrementRepository)

	documentContent, err := documentService.GenerateContent(ctx, text)
	if err != nil {
		return domain.Increment{}, err
	}

	incrementHandle, err := incrementService.IssueHandle()
	if err != nil {
		return domain.Increment{}, err
	}

	increment := domain.NewIncrement(incrementHandle, previousIncrementHandle, []domain.DocumentChange{})
	_, err = incrementService.Create(increment)
	if err != nil {
		return domain.Increment{}, err
	}

	documentChange := domain.NewDocumentCreateChange(documentContent)
	increment, err = incrementService.AddDocumentChange(increment, documentChange)
	if err != nil {
		return domain.Increment{}, err
	}

	return increment, err
}
