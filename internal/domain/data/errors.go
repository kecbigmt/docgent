package data

import "errors"

var (
	// ErrFileNotFound はファイルが見つからない場合のエラー
	ErrFileNotFound = errors.New("file not found")

	// ErrFileAlreadyExists は作成しようとしたファイルが既に存在する場合のエラー
	ErrFileAlreadyExists = errors.New("file already exists")

	// ErrInvalidFrontmatter はフロントマターの形式が不正な場合のエラー
	ErrInvalidFrontmatter = errors.New("invalid frontmatter format")

	// ErrInvalidKnowledgeSource は知識源の形式が不正な場合のエラー
	ErrInvalidKnowledgeSource = errors.New("invalid knowledge source format")

	// ErrFailedToAccessFile はファイルへのアクセスに失敗した場合のエラー
	ErrFailedToAccessFile = errors.New("failed to access file")
)
