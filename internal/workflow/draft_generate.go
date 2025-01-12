package workflow

import (
	"context"

	"docgent-backend/internal/model"
	"docgent-backend/internal/model/infrastructure"
)

type DraftGenerateWorkflow struct {
	documentationAgent infrastructure.DocumentationAgent
	documentStore      infrastructure.DocumentStore
}

func NewDraftGenerateWorkflow(
	documentationAgent infrastructure.DocumentationAgent,
	documentStore infrastructure.DocumentStore,
) *DraftGenerateWorkflow {
	return &DraftGenerateWorkflow{
		documentationAgent: documentationAgent,
		documentStore:      documentStore,
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
