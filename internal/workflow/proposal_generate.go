package workflow

import (
	"context"
	"fmt"
	"strings"

	"docgent-backend/internal/domain"
)

type ProposalGenerateWorkflow struct {
	documentAgent      domain.DocumentAgent
	proposalAgent      domain.ProposalAgent
	proposalRepository domain.ProposalRepository
}

func NewProposalGenerateWorkflow(
	documentAgent domain.DocumentAgent,
	proposalAgent domain.ProposalAgent,
	proposalRepository domain.ProposalRepository,
) *ProposalGenerateWorkflow {
	return &ProposalGenerateWorkflow{
		documentAgent:      documentAgent,
		proposalAgent:      proposalAgent,
		proposalRepository: proposalRepository,
	}
}

func (w *ProposalGenerateWorkflow) Execute(
	ctx context.Context,
	text string,
) (domain.ProposalHandle, error) {
	documentService := domain.NewDocumentService(w.documentAgent)

	documentContent, err := documentService.GenerateContent(ctx, text)
	if err != nil {
		return domain.ProposalHandle{}, err
	}

	diffBody := fmt.Sprintf("@@ -0,0 +1,%d @@\n", strings.Count(documentContent.Body, "\n"))
	for _, line := range strings.Split(documentContent.Body, "\n") {
		diffBody += "+" + line + "\n"
	}

	diff := domain.NewCreateDiff(documentContent.Title+".md", diffBody)
	diffs := domain.Diffs([]domain.Diff{diff})

	contextDescription := fmt.Sprintf("以下の指示に従ってドキュメントを生成しました。\n\n%s", text)

	// Create proposal using the increment
	proposalService := domain.NewProposalService(w.proposalAgent, w.proposalRepository)
	proposalContent, err := proposalService.GenerateContent(diffs, contextDescription)
	if err != nil {
		return domain.ProposalHandle{}, fmt.Errorf("failed to generate proposal content: %w", err)
	}

	proposal, err := proposalService.Create(diffs, proposalContent)
	if err != nil {
		return domain.ProposalHandle{}, fmt.Errorf("failed to create proposal: %w", err)
	}

	return proposal, nil
}
