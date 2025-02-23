package data

import "errors"

var (
	// ErrFileNotFound はファイルが見つからない場合のエラー
	ErrFileNotFound = errors.New("file not found")
	// ErrFileUpdateFailed はファイルの更新に失敗した場合のエラー
	ErrFileUpdateFailed = errors.New("failed to update file")
)
