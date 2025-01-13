package workflow

import (
	"context"
	"fmt"

	"docgent-backend/internal/domain"
)

type ProposalGenerateWorkflow struct {
	documentAgent       domain.DocumentAgent
	incrementRepository domain.IncrementRepository
	proposalAgent       domain.ProposalAgent
	proposalRepository  domain.ProposalRepository
}

func NewProposalGenerateWorkflow(
	documentAgent domain.DocumentAgent,
	incrementRepository domain.IncrementRepository,
	proposalAgent domain.ProposalAgent,
	proposalRepository domain.ProposalRepository,
) *ProposalGenerateWorkflow {
	return &ProposalGenerateWorkflow{
		documentAgent:       documentAgent,
		incrementRepository: incrementRepository,
		proposalAgent:       proposalAgent,
		proposalRepository:  proposalRepository,
	}
}

func (w *ProposalGenerateWorkflow) Execute(
	ctx context.Context,
	text string,
	previousIncrementHandle domain.IncrementHandle,
) (domain.Proposal, error) {
	documentService := domain.NewDocumentService(w.documentAgent)
	incrementService := domain.NewIncrementService(w.incrementRepository)

	documentContent, err := documentService.GenerateContent(ctx, text)
	if err != nil {
		return domain.Proposal{}, err
	}

	incrementHandle, err := incrementService.IssueHandle()
	if err != nil {
		return domain.Proposal{}, err
	}

	increment := domain.NewIncrement(incrementHandle, previousIncrementHandle, []domain.DocumentChange{})
	_, err = incrementService.Create(increment)
	if err != nil {
		return domain.Proposal{}, err
	}

	documentChange := domain.NewDocumentCreateChange(documentContent)
	increment, err = incrementService.AddDocumentChange(increment, documentChange)
	if err != nil {
		return domain.Proposal{}, err
	}

	// Create proposal using the increment
	proposalService := domain.NewProposalService(w.proposalAgent, w.proposalRepository)
	proposalContent, err := proposalService.GenerateContent(increment)
	if err != nil {
		return domain.Proposal{}, fmt.Errorf("failed to generate proposal content: %w", err)
	}

	proposal, err := proposalService.Create(proposalContent, increment)
	if err != nil {
		return domain.Proposal{}, fmt.Errorf("failed to create proposal: %w", err)
	}

	return proposal, nil
}
