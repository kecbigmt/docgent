package infrastructure

import (
	"context"
)

type DocumentDraft struct {
	Title   string
	Content string
}

type GenerativeModel interface {
	GenerateDocument(ctx context.Context, input string) (DocumentDraft, error)
}
