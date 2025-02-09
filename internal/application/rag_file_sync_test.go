package application

import (
	"errors"
	"testing"

	"docgent-backend/internal/application/port"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestRagFileSyncUsecase_Execute(t *testing.T) {
	tests := []struct {
		name          string
		newFiles      []string
		modifiedFiles []string
		deletedFiles  []string
		setupMocks    func(*MockRAGCorpus, *MockFileQueryService)
		expectedError error
	}{
		{
			name:          "正常系：新規ファイルのアップロード",
			newFiles:      []string{"docs/new.md"},
			modifiedFiles: []string{},
			deletedFiles:  []string{},
			setupMocks: func(ragCorpus *MockRAGCorpus, fileQueryService *MockFileQueryService) {
				ragCorpus.On("ListFiles", mock.Anything).Return([]port.RAGFile{}, nil)
				fileQueryService.On("FindFile", mock.Anything, "docs/new.md").Return(port.File{
					Path:    "docs/new.md",
					Content: "新規ファイルの内容",
				}, nil)
				ragCorpus.On("UploadFile", mock.Anything, mock.Anything, "docs/new.md", mock.Anything).Return(nil)
			},
			expectedError: nil,
		},
		{
			name:          "正常系：既存ファイルの更新",
			newFiles:      []string{},
			modifiedFiles: []string{"docs/modified.md"},
			deletedFiles:  []string{},
			setupMocks: func(ragCorpus *MockRAGCorpus, fileQueryService *MockFileQueryService) {
				ragCorpus.On("ListFiles", mock.Anything).Return([]port.RAGFile{
					{ID: 1, FileName: "docs/modified.md"},
				}, nil)
				fileQueryService.On("FindFile", mock.Anything, "docs/modified.md").Return(port.File{
					Path:    "docs/modified.md",
					Content: "更新されたファイルの内容",
				}, nil)
				ragCorpus.On("UploadFile", mock.Anything, mock.Anything, "docs/modified.md", mock.Anything).Return(nil)
				ragCorpus.On("DeleteFile", mock.Anything, int64(1)).Return(nil)
			},
			expectedError: nil,
		},
		{
			name:          "正常系：ファイルの削除",
			newFiles:      []string{},
			modifiedFiles: []string{},
			deletedFiles:  []string{"docs/deleted.md"},
			setupMocks: func(ragCorpus *MockRAGCorpus, fileQueryService *MockFileQueryService) {
				ragCorpus.On("ListFiles", mock.Anything).Return([]port.RAGFile{
					{ID: 1, FileName: "docs/deleted.md"},
				}, nil)
				ragCorpus.On("DeleteFile", mock.Anything, int64(1)).Return(nil)
			},
			expectedError: nil,
		},
		{
			name:          "エラー系：ListFilesに失敗",
			newFiles:      []string{"docs/new.md"},
			modifiedFiles: []string{},
			deletedFiles:  []string{},
			setupMocks: func(ragCorpus *MockRAGCorpus, fileQueryService *MockFileQueryService) {
				ragCorpus.On("ListFiles", mock.Anything).Return([]port.RAGFile{}, errors.New("failed to list files"))
			},
			expectedError: errors.New("failed to list files"),
		},
		{
			name:          "エラー系：FindFileに失敗",
			newFiles:      []string{"docs/new.md"},
			modifiedFiles: []string{},
			deletedFiles:  []string{},
			setupMocks: func(ragCorpus *MockRAGCorpus, fileQueryService *MockFileQueryService) {
				ragCorpus.On("ListFiles", mock.Anything).Return([]port.RAGFile{}, nil)
				fileQueryService.On("FindFile", mock.Anything, "docs/new.md").Return(port.File{}, errors.New("failed to find file"))
			},
			expectedError: errors.New("failed to find file"),
		},
		{
			name:          "エラー系：UploadFileに失敗",
			newFiles:      []string{"docs/new.md"},
			modifiedFiles: []string{},
			deletedFiles:  []string{},
			setupMocks: func(ragCorpus *MockRAGCorpus, fileQueryService *MockFileQueryService) {
				ragCorpus.On("ListFiles", mock.Anything).Return([]port.RAGFile{}, nil)
				fileQueryService.On("FindFile", mock.Anything, "docs/new.md").Return(port.File{
					Path:    "docs/new.md",
					Content: "新規ファイルの内容",
				}, nil)
				ragCorpus.On("UploadFile", mock.Anything, mock.Anything, "docs/new.md", mock.Anything).Return(errors.New("failed to upload file"))
			},
			expectedError: errors.New("failed to upload file"),
		},
		{
			name:          "エラー系：DeleteFileに失敗",
			newFiles:      []string{},
			modifiedFiles: []string{},
			deletedFiles:  []string{"docs/deleted.md"},
			setupMocks: func(ragCorpus *MockRAGCorpus, fileQueryService *MockFileQueryService) {
				ragCorpus.On("ListFiles", mock.Anything).Return([]port.RAGFile{
					{ID: 1, FileName: "docs/deleted.md"},
				}, nil)
				ragCorpus.On("DeleteFile", mock.Anything, int64(1)).Return(errors.New("failed to delete file"))
			},
			expectedError: errors.New("failed to delete file"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// モックの準備
			ragCorpus := new(MockRAGCorpus)
			fileQueryService := new(MockFileQueryService)

			tt.setupMocks(ragCorpus, fileQueryService)

			// ユースケースの作成
			usecase := NewRagFileSyncUsecase(ragCorpus, fileQueryService)

			// テストの実行
			err := usecase.Execute(tt.newFiles, tt.modifiedFiles, tt.deletedFiles)

			// アサーション
			if tt.expectedError != nil {
				assert.Error(t, err)
				assert.Equal(t, tt.expectedError.Error(), err.Error())
			} else {
				assert.NoError(t, err)
			}

			// モックの検証
			ragCorpus.AssertExpectations(t)
			fileQueryService.AssertExpectations(t)
		})
	}
}
