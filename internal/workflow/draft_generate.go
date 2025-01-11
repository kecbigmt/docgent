package workflow

import (
	"context"

	"docgent-backend/internal/model"
	"docgent-backend/internal/model/infrastructure"
)

type DraftGenerateWorkflowParams struct {
	DocumentationAgent infrastructure.DocumentationAgent
}

type DraftGenerateWorkflow struct {
	documentationAgent infrastructure.DocumentationAgent
}

func NewDraftGenerateWorkflow(p DraftGenerateWorkflowParams) *DraftGenerateWorkflow {
	return &DraftGenerateWorkflow{
		documentationAgent: p.DocumentationAgent,
	}
}

func (w *DraftGenerateWorkflow) Execute(ctx context.Context, text string) (*model.Draft, error) {
	rawDraft, err := w.documentationAgent.GenerateDocumentDraft(ctx, text)
	if err != nil {
		return nil, err
	}

	draft, err := model.NewDraft(rawDraft.Title, rawDraft.Content)
	if err != nil {
		return nil, err
	}

	return &draft, nil
}
