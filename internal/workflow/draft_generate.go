package workflow

import (
	"context"

	"docgent-backend/internal/model"
	"docgent-backend/internal/model/infrastructure"
)

type DraftGenerateWorkflowParams struct {
	GenerativeModel infrastructure.GenerativeModel
}

type DraftGenerateWorkflow struct {
	generativeModel infrastructure.GenerativeModel
}

func NewDraftGenerateWorkflow(p DraftGenerateWorkflowParams) *DraftGenerateWorkflow {
	return &DraftGenerateWorkflow{
		generativeModel: p.GenerativeModel,
	}
}

func (w *DraftGenerateWorkflow) Execute(ctx context.Context, text string) (*model.Draft, error) {
	rawDraft, err := w.generativeModel.GenerateDocument(ctx, text)
	if err != nil {
		return nil, err
	}

	draft, err := model.NewDraft(rawDraft.Title, rawDraft.Content)
	if err != nil {
		return nil, err
	}

	return &draft, nil
}
