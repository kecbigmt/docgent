package model

type Draft struct {
	ID      string
	Title   string
	Content string
}

func NewDraft(id, title, content string) (Draft, error) {
	return Draft{
		ID:      id,
		Title:   title,
		Content: content,
	}, nil
}
