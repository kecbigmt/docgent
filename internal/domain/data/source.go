package data

import "context"

type Source struct {
	uri     *URI
	content string
}

func NewSource(uri *URI, content string) *Source {
	return &Source{uri: uri, content: content}
}

func (s *Source) URI() *URI {
	return s.uri
}

func (s *Source) Content() string {
	return s.content
}

type SourceRepository interface {
	Find(ctx context.Context, uri *URI) (*Source, error)
}
