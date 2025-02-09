package rag

import (
	"context"
	"docgent-backend/internal/application/port"
	"docgent-backend/internal/infrastructure/google/vertexai/rag/lib"
	"encoding/json"
	"io"
	"net/http"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCorpus_UploadFile(t *testing.T) {
	tests := []struct {
		name          string
		file          io.Reader
		fileName      string
		options       []port.RAGCorpusUploadFileOption
		setup         func(*mockTransport)
		errorExpected bool
		expectedReqs  []mockRequest
	}{
		{
			name:     "正常系: ファイルのアップロードが成功する場合",
			file:     strings.NewReader("test content"),
			fileName: "test.md",
			options: []port.RAGCorpusUploadFileOption{
				func(o *port.RAGCorpusUploadFileOptions) {
					o.Description = "test description"
					o.ChunkingConfig = port.ChunkingConfig{
						ChunkSize:    100,
						ChunkOverlap: 10,
					}
				},
			},
			setup: func(mt *mockTransport) {
				mt.responses = map[string]mockResponse{
					"POST /upload/v1/projects/test-project/locations/test-location/ragCorpora/1/ragFiles:upload": {
						statusCode: http.StatusOK,
						body:       map[string]interface{}{},
					},
				}
			},
			errorExpected: false,
			expectedReqs: []mockRequest{
				{
					method: "POST",
					path:   "/upload/v1/projects/test-project/locations/test-location/ragCorpora/1/ragFiles:upload",
					validateMultipartForm: func(t *testing.T, r *http.Request) {
						err := r.ParseMultipartForm(32 << 20)
						assert.NoError(t, err)

						// メタデータの検証
						metadata := r.FormValue("metadata")
						var metadataObj lib.UploadFileMetadata
						err = json.Unmarshal([]byte(metadata), &metadataObj)
						assert.NoError(t, err)

						assert.Equal(t, "test.md", metadataObj.RagFile.DisplayName)
						assert.Equal(t, "test description", metadataObj.RagFile.Description)
						assert.NotNil(t, metadataObj.UploadRagFileConfig)
						assert.NotNil(t, metadataObj.UploadRagFileConfig.RagFileTransformationConfig)
						assert.NotNil(t, metadataObj.UploadRagFileConfig.RagFileTransformationConfig.RagFileChunkingConfig)
						assert.Equal(t, int32(100), metadataObj.UploadRagFileConfig.RagFileTransformationConfig.RagFileChunkingConfig.ChunkSize)
						assert.Equal(t, int32(10), metadataObj.UploadRagFileConfig.RagFileTransformationConfig.RagFileChunkingConfig.ChunkOverlap)

						// ファイルの検証
						file, header, err := r.FormFile("file")
						assert.NoError(t, err)
						defer file.Close()

						assert.Equal(t, "test.md", header.Filename)
						content, err := io.ReadAll(file)
						assert.NoError(t, err)
						assert.Equal(t, "test content", string(content))
					},
				},
			},
		},
		{
			name:     "正常系: オプションなしでファイルのアップロードが成功する場合",
			file:     strings.NewReader("test content"),
			fileName: "test.md",
			options:  []port.RAGCorpusUploadFileOption{},
			setup: func(mt *mockTransport) {
				mt.responses = map[string]mockResponse{
					"POST /upload/v1/projects/test-project/locations/test-location/ragCorpora/1/ragFiles:upload": {
						statusCode: http.StatusOK,
						body:       map[string]interface{}{},
					},
				}
			},
			errorExpected: false,
			expectedReqs: []mockRequest{
				{
					method: "POST",
					path:   "/upload/v1/projects/test-project/locations/test-location/ragCorpora/1/ragFiles:upload",
					validateMultipartForm: func(t *testing.T, r *http.Request) {
						err := r.ParseMultipartForm(32 << 20)
						assert.NoError(t, err)

						// メタデータの検証
						metadata := r.FormValue("metadata")
						var metadataObj lib.UploadFileMetadata
						err = json.Unmarshal([]byte(metadata), &metadataObj)
						assert.NoError(t, err)

						assert.Equal(t, "test.md", metadataObj.RagFile.DisplayName)
						assert.Empty(t, metadataObj.RagFile.Description)
						assert.Nil(t, metadataObj.UploadRagFileConfig)

						// ファイルの検証
						file, header, err := r.FormFile("file")
						assert.NoError(t, err)
						defer file.Close()

						assert.Equal(t, "test.md", header.Filename)
						content, err := io.ReadAll(file)
						assert.NoError(t, err)
						assert.Equal(t, "test content", string(content))
					},
				},
			},
		},
		{
			name:     "異常系: APIエラーの場合",
			file:     strings.NewReader("test content"),
			fileName: "test.md",
			options:  []port.RAGCorpusUploadFileOption{},
			setup: func(mt *mockTransport) {
				mt.responses = map[string]mockResponse{
					"POST /upload/v1/projects/test-project/locations/test-location/ragCorpora/1/ragFiles:upload": {
						statusCode: http.StatusInternalServerError,
						body:       map[string]interface{}{},
					},
				}
			},
			errorExpected: true,
			expectedReqs: []mockRequest{
				{
					method: "POST",
					path:   "/upload/v1/projects/test-project/locations/test-location/ragCorpora/1/ragFiles:upload",
					validateMultipartForm: func(t *testing.T, r *http.Request) {
						err := r.ParseMultipartForm(32 << 20)
						assert.NoError(t, err)
					},
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// モックトランスポーターの準備
			mt := newMockTransport(t, tt.expectedReqs)
			tt.setup(mt)

			// テスト対象のインスタンスを作成
			client := lib.NewClient(&http.Client{Transport: mt}, "test-project", "test-location")
			corpus := NewCorpus(client, 1)

			// テスト実行
			err := corpus.UploadFile(
				context.Background(),
				tt.file,
				tt.fileName,
				tt.options...,
			)

			// アサーション
			if tt.errorExpected {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

			// リクエストの検証
			mt.verify(t)
		})
	}
}
