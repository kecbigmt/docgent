package workflow

import (
	"context"

	"docgent-backend/internal/model"
	"docgent-backend/internal/model/infrastructure"
)

type DraftGenerateWorkflowParams struct {
	DocumentationAgent infrastructure.DocumentationAgent
	DocumentStore      infrastructure.DocumentStore
}

type DraftGenerateWorkflow struct {
	documentationAgent infrastructure.DocumentationAgent
	documentStore      infrastructure.DocumentStore
}

func NewDraftGenerateWorkflow(p DraftGenerateWorkflowParams) *DraftGenerateWorkflow {
	return &DraftGenerateWorkflow{
		documentationAgent: p.DocumentationAgent,
		documentStore:      p.DocumentStore,
	}
}

func (w *DraftGenerateWorkflow) Execute(ctx context.Context, text string) (*model.Draft, error) {
	rawDraft, err := w.documentationAgent.GenerateDocumentDraft(ctx, text)
	if err != nil {
		return nil, err
	}

	// Save the draft using DocumentStore
	savedDoc, err := w.documentStore.Save(infrastructure.DocumentInput(rawDraft))
	if err != nil {
		return nil, err
	}

	draft, err := model.NewDraft(savedDoc.ID, savedDoc.Title, savedDoc.Content)
	if err != nil {
		return nil, err
	}

	return &draft, nil
}
