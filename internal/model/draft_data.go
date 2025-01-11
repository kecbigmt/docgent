package model

type Draft struct {
	Title   string
	Content string
}

func NewDraft(title, content string) (Draft, error) {
	return Draft{Title: title, Content: content}, nil
}
