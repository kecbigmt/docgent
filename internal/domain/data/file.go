package data

import "context"

// File はドキュメントファイルを表す
type File struct {
	Path       string
	Content    string
	SourceURIs []URI
}

// FileRepository はファイルの永続化を担当
type FileRepository interface {
	// Create は新しいファイルを作成する
	Create(ctx context.Context, file *File) error
	// Update は既存のファイルを更新する
	Update(ctx context.Context, file *File) error
	// Get は指定されたパスのファイルを取得する
	Get(ctx context.Context, path string) (*File, error)
	// Delete は指定されたパスのファイルを削除する
	Delete(ctx context.Context, path string) error
}
