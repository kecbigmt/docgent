package infrastructure

type DocumentInput struct {
	Title   string
	Content string
}

type Document struct {
	ID      string
	Title   string
	Content string
}

type DocumentStore interface {
	Save(documentInput DocumentInput) (Document, error)
}
