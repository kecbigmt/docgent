package infrastructure

import "context"

type GenerativeModel interface {
	GenerateDocument(ctx context.Context, input string) (string, error)
}
