package domain

/**
 * Document
 */

type Document struct {
	Handle DocumentHandle
	DocumentContent
}

func NewDocument(handle DocumentHandle, content DocumentContent) Document {
	return Document{
		Handle:          handle,
		DocumentContent: content,
	}
}

/**
 * DocumentChange
 */

type DocumentChangeType int

const (
	DocumentCreateChange = iota
	DocumentUpdateChange
	DocumentDeleteChange
)

type DocumentChange struct {
	Type           DocumentChangeType
	DocumentHandle DocumentHandle
	DocumentContent
}

type DocumentChangeOption func(*DocumentChange)

func NewDocumentCreateChange(newContent DocumentContent) DocumentChange {
	return DocumentChange{
		Type:            DocumentCreateChange,
		DocumentContent: newContent,
	}
}

func NewDocumentUpdateChange(handle DocumentHandle, newContent DocumentContent) DocumentChange {
	return DocumentChange{
		Type:            DocumentUpdateChange,
		DocumentHandle:  handle,
		DocumentContent: newContent,
	}
}

func NewDocumentDeleteChange(handle DocumentHandle) DocumentChange {
	return DocumentChange{
		Type:           DocumentDeleteChange,
		DocumentHandle: handle,
	}
}

/**
 * DocumentHandle
 */

type DocumentHandle struct {
	Source string
	Value  string
}

func NewDocumentHandle(source, value string) DocumentHandle {
	return DocumentHandle{
		Source: source,
		Value:  value,
	}
}

/**
 * DocumentContent
 */

type DocumentContent struct {
	Title string
	Body  string
}

func NewDocumentContent(title, body string) DocumentContent {
	return DocumentContent{
		Title: title,
		Body:  body,
	}
}
