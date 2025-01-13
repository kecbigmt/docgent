package workflow

import (
	"context"

	"docgent-backend/internal/domain"
)

type DraftGenerateWorkflowParams struct {
	DocumentationAgent domain.DocumentationAgent
	DocumentStore      domain.DocumentStore
}

type DraftGenerateWorkflow struct {
	documentationAgent domain.DocumentationAgent
	documentStore      domain.DocumentStore
}

func NewDraftGenerateWorkflow(params DraftGenerateWorkflowParams) *DraftGenerateWorkflow {
	return &DraftGenerateWorkflow{
		documentationAgent: params.DocumentationAgent,
		documentStore:      params.DocumentStore,
	}
}

func (w *DraftGenerateWorkflow) Execute(ctx context.Context, text string) (domain.Draft, error) {
	rawDraft, err := w.documentationAgent.GenerateDocumentDraft(ctx, text)
	if err != nil {
		return domain.Draft{}, err
	}

	// Save the draft using DocumentStore
	savedDoc, err := w.documentStore.Save(domain.DocumentInput(rawDraft))
	if err != nil {
		return domain.Draft{}, err
	}

	draft, err := domain.NewDraft(savedDoc.ID, savedDoc.Title, savedDoc.Content)
	if err != nil {
		return domain.Draft{}, err
	}

	return draft, nil
}
