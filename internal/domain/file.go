package domain

type File struct {
	Name    string
	Content string
}

type FileQueryService interface {
	Find(name string) (File, error)
}
