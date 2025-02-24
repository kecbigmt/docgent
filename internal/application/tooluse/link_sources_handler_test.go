package tooluse

import (
	"context"
	"testing"

	"docgent/internal/domain/data"
	"docgent/internal/domain/tooluse"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type MockFileRepository struct {
	mock.Mock
}

func (m *MockFileRepository) Create(ctx context.Context, file *data.File) error {
	args := m.Called(ctx, file)
	return args.Error(0)
}

func (m *MockFileRepository) Get(ctx context.Context, path string) (*data.File, error) {
	args := m.Called(ctx, path)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*data.File), args.Error(1)
}

func (m *MockFileRepository) Update(ctx context.Context, file *data.File) error {
	args := m.Called(ctx, file)
	return args.Error(0)
}

func (m *MockFileRepository) Delete(ctx context.Context, path string) error {
	args := m.Called(ctx, path)
	return args.Error(0)
}

func TestLinkSourcesHandler_Handle(t *testing.T) {
	tests := []struct {
		name           string
		toolUse        tooluse.LinkSources
		setupMocks     func(*MockFileRepository)
		expectedResult string
		expectedError  error
	}{
		{
			name: "正常系：新しい知識源を追加",
			toolUse: tooluse.NewLinkSources(
				"path/to/file.md",
				[]string{"https://github.com/user/repo/pull/1"},
			),
			setupMocks: func(fileRepository *MockFileRepository) {
				existingFile := &data.File{
					Path:    "path/to/file.md",
					Content: "# Hello\nWorld",
					SourceURIs: []*data.URI{
						data.NewURIUnsafe("https://slack.com/archives/C01234567/p123456789"),
					},
				}
				expectedFile := &data.File{
					Path:    "path/to/file.md",
					Content: "# Hello\nWorld",
					SourceURIs: []*data.URI{
						data.NewURIUnsafe("https://slack.com/archives/C01234567/p123456789"),
						data.NewURIUnsafe("https://github.com/user/repo/pull/1"),
					},
				}
				fileRepository.On("Get", mock.Anything, "path/to/file.md").Return(existingFile, nil)
				fileRepository.On("Update", mock.Anything, mock.MatchedBy(func(file *data.File) bool {
					return file.Path == expectedFile.Path &&
						file.Content == expectedFile.Content &&
						len(file.SourceURIs) == len(expectedFile.SourceURIs) &&
						file.SourceURIs[0].Equal(expectedFile.SourceURIs[0]) &&
						file.SourceURIs[1].Equal(expectedFile.SourceURIs[1])
				})).Return(nil)
			},
			expectedResult: "<success>Knowledge sources added</success>",
			expectedError:  nil,
		},
		{
			name: "正常系：重複する知識源は追加しない",
			toolUse: tooluse.NewLinkSources(
				"path/to/file.md",
				[]string{"https://slack.com/archives/C01234567/p123456789"},
			),
			setupMocks: func(fileRepository *MockFileRepository) {
				existingFile := &data.File{
					Path:    "path/to/file.md",
					Content: "# Hello\nWorld",
					SourceURIs: []*data.URI{
						data.NewURIUnsafe("https://slack.com/archives/C01234567/p123456789"),
					},
				}
				expectedFile := &data.File{
					Path:    "path/to/file.md",
					Content: "# Hello\nWorld",
					SourceURIs: []*data.URI{
						data.NewURIUnsafe("https://slack.com/archives/C01234567/p123456789"),
					},
				}
				fileRepository.On("Get", mock.Anything, "path/to/file.md").Return(existingFile, nil)
				fileRepository.On("Update", mock.Anything, mock.MatchedBy(func(file *data.File) bool {
					return file.Path == expectedFile.Path &&
						file.Content == expectedFile.Content &&
						len(file.SourceURIs) == len(expectedFile.SourceURIs) &&
						file.SourceURIs[0].Equal(expectedFile.SourceURIs[0])
				})).Return(nil)
			},
			expectedResult: "<success>Knowledge sources added</success>",
			expectedError:  nil,
		},
		{
			name: "エラー系：ファイルの取得に失敗",
			toolUse: tooluse.NewLinkSources(
				"path/to/file.md",
				[]string{"https://github.com/user/repo/pull/1"},
			),
			setupMocks: func(fileRepository *MockFileRepository) {
				fileRepository.On("Get", mock.Anything, "path/to/file.md").Return(nil, data.ErrFileNotFound)
			},
			expectedResult: "",
			expectedError:  data.ErrFileNotFound,
		},
		{
			name: "エラー系：ファイルの更新に失敗",
			toolUse: tooluse.NewLinkSources(
				"path/to/file.md",
				[]string{"https://github.com/user/repo/pull/1"},
			),
			setupMocks: func(fileRepository *MockFileRepository) {
				existingFile := &data.File{
					Path:    "path/to/file.md",
					Content: "# Hello\nWorld",
					SourceURIs: []*data.URI{
						data.NewURIUnsafe("https://slack.com/archives/C01234567/p123456789"),
					},
				}
				expectedFile := &data.File{
					Path:    "path/to/file.md",
					Content: "# Hello\nWorld",
					SourceURIs: []*data.URI{
						data.NewURIUnsafe("https://slack.com/archives/C01234567/p123456789"),
						data.NewURIUnsafe("https://github.com/user/repo/pull/1"),
					},
				}
				fileRepository.On("Get", mock.Anything, "path/to/file.md").Return(existingFile, nil)
				fileRepository.On("Update", mock.Anything, mock.MatchedBy(func(file *data.File) bool {
					return file.Path == expectedFile.Path &&
						file.Content == expectedFile.Content &&
						len(file.SourceURIs) == len(expectedFile.SourceURIs) &&
						file.SourceURIs[0].Equal(expectedFile.SourceURIs[0]) &&
						file.SourceURIs[1].Equal(expectedFile.SourceURIs[1])
				})).Return(data.ErrFailedToAccessFile)
			},
			expectedResult: "",
			expectedError:  data.ErrFailedToAccessFile,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// モックの準備
			fileRepository := new(MockFileRepository)
			tt.setupMocks(fileRepository)

			fileChanged := false
			handler := NewLinkSourcesHandler(context.Background(), fileRepository, &fileChanged)

			// テストの実行
			result, _, err := handler.Handle(tt.toolUse)

			// アサーション
			if tt.expectedError != nil {
				assert.Error(t, err)
				assert.Equal(t, tt.expectedError, err)
				assert.False(t, fileChanged)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedResult, result)
				assert.True(t, fileChanged)
			}

			// モックの検証
			fileRepository.AssertExpectations(t)
		})
	}
}
