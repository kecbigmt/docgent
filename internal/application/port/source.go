package port

import (
	"context"
	"docgent/internal/domain/data"
	"errors"
)

var ErrUnsupportedSource = errors.New("unsupported source")

type SourceRepository interface {
	Match(uri *data.URI) bool
	data.SourceRepository
}

type SourceRepositoryManager struct {
	sourceRepositories []SourceRepository
}

func NewSourceRepositoryManager(sourceRepositories []SourceRepository) *SourceRepositoryManager {
	return &SourceRepositoryManager{sourceRepositories: sourceRepositories}
}

func (m *SourceRepositoryManager) Find(ctx context.Context, uri *data.URI) (*data.Source, error) {
	for _, sourceRepository := range m.sourceRepositories {
		if sourceRepository.Match(uri) {
			return sourceRepository.Find(ctx, uri)
		}
	}
	return nil, ErrUnsupportedSource
}
