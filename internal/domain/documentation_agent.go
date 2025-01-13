package domain

import (
	"context"
)

type DocumentDraft struct {
	Title   string
	Content string
}

type DocumentationAgent interface {
	GenerateDocumentDraft(ctx context.Context, input string) (DocumentDraft, error)
}
