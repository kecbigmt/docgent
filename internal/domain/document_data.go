package domain

type Document struct {
	Title   string
	Content string
}

func NewDocument(title, content string) Document {
	return Document{
		Title:   title,
		Content: content,
	}
}
