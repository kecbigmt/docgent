package workflow

import (
	"context"
	"fmt"

	"docgent-backend/internal/domain"
)

type ProposalCommentWorkflow struct {
	proposalAgent      domain.ProposalAgent
	proposalRepository domain.ProposalRepository
}

func NewProposalCommentWorkflow(
	proposalAgent domain.ProposalAgent,
	proposalRepository domain.ProposalRepository,
) *ProposalCommentWorkflow {
	return &ProposalCommentWorkflow{
		proposalAgent:      proposalAgent,
		proposalRepository: proposalRepository,
	}
}

func (w *ProposalCommentWorkflow) Execute(
	ctx context.Context,
	proposalHandle domain.ProposalHandle,
	commentBody string,
) (domain.Comment, error) {
	proposalService := domain.NewProposalService(w.proposalAgent, w.proposalRepository)

	comment, err := proposalService.CreateComment(proposalHandle, commentBody)
	if err != nil {
		return domain.Comment{}, fmt.Errorf("failed to add comment to proposal: %w", err)
	}

	return comment, nil
}
