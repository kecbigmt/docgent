package application

import (
	"context"
	"fmt"

	"docgent-backend/internal/domain"
)

type ProposalCommentUsecase struct {
	proposalAgent      domain.ProposalAgent
	proposalRepository domain.ProposalRepository
}

func NewProposalCommentUsecase(
	proposalAgent domain.ProposalAgent,
	proposalRepository domain.ProposalRepository,
) *ProposalCommentUsecase {
	return &ProposalCommentUsecase{
		proposalAgent:      proposalAgent,
		proposalRepository: proposalRepository,
	}
}

func (w *ProposalCommentUsecase) Execute(
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
