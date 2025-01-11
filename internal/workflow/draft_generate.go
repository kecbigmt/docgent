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
	doc, err := w.generativeModel.GenerateDocument(ctx, text)
	if err != nil {
		return &model.Draft{}, err
	}

	draft, err := model.NewDraft("new draft", doc)
	if err != nil {
		return &model.Draft{}, err
	}

	return &draft, nil
}
