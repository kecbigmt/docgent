package lib

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestClient_UploadFile(t *testing.T) {
	tests := []struct {
		name            string
		file            io.Reader
		fileName        string
		options         []UploadFileOption
		setup           func(*mockTransport)
		errorExpected   bool
		expectedReqs    []mockRequest
		expectedRagFile File
	}{
		{
			name:     "Success: Upload file with options",
			file:     strings.NewReader("test content"),
			fileName: "test.md",
			options: []UploadFileOption{
				WithUploadFileDescription("test description"),
				WithUploadFileChunkingConfig(100, 10),
			},
			setup: func(mt *mockTransport) {
				mt.responses = map[string]mockResponse{
					"POST /upload/v1/projects/test-project/locations/test-location/ragCorpora/1/ragFiles:upload": {
						statusCode: http.StatusOK,
						body: map[string]interface{}{
							"ragFile": map[string]interface{}{
								"name":        "test-project/test-location/ragCorpora/1/ragFiles/123",
								"displayName": "test.md",
								"description": "test description",
							},
						},
					},
				}
			},
			expectedRagFile: File{
				Name:        "test-project/test-location/ragCorpora/1/ragFiles/123",
				DisplayName: "test.md",
				Description: "test description",
			},
			errorExpected: false,
			expectedReqs: []mockRequest{
				{
					method: "POST",
					path:   "/upload/v1/projects/test-project/locations/test-location/ragCorpora/1/ragFiles:upload",
					validateMultipartForm: func(t *testing.T, r *http.Request) {
						err := r.ParseMultipartForm(32 << 20)
						assert.NoError(t, err)

						// Validate metadata
						metadata := r.FormValue("metadata")
						var metadataObj UploadFileMetadata
						err = json.Unmarshal([]byte(metadata), &metadataObj)
						assert.NoError(t, err)

						assert.Equal(t, "test.md", metadataObj.RagFile.DisplayName)
						assert.Equal(t, "test description", metadataObj.RagFile.Description)
						assert.NotNil(t, metadataObj.UploadRagFileConfig)
						assert.NotNil(t, metadataObj.UploadRagFileConfig.RagFileTransformationConfig)
						assert.NotNil(t, metadataObj.UploadRagFileConfig.RagFileTransformationConfig.RagFileChunkingConfig)
						assert.Equal(t, int32(100), metadataObj.UploadRagFileConfig.RagFileTransformationConfig.RagFileChunkingConfig.FixedLengthChunking.ChunkSize)
						assert.Equal(t, int32(10), metadataObj.UploadRagFileConfig.RagFileTransformationConfig.RagFileChunkingConfig.FixedLengthChunking.ChunkOverlap)

						// Validate file
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
			name:     "Success: Upload file without options",
			file:     strings.NewReader("test content"),
			fileName: "test.md",
			options:  []UploadFileOption{},
			setup: func(mt *mockTransport) {
				mt.responses = map[string]mockResponse{
					"POST /upload/v1/projects/test-project/locations/test-location/ragCorpora/1/ragFiles:upload": {
						statusCode: http.StatusOK,
						body: map[string]interface{}{
							"ragFile": map[string]interface{}{
								"name":        "test-project/test-location/ragCorpora/1/ragFiles/123",
								"displayName": "test.md",
								"description": "",
							},
						},
					},
				}
			},
			errorExpected: false,
			expectedRagFile: File{
				Name:        "test-project/test-location/ragCorpora/1/ragFiles/123",
				DisplayName: "test.md",
				Description: "",
			},
			expectedReqs: []mockRequest{
				{
					method: "POST",
					path:   "/upload/v1/projects/test-project/locations/test-location/ragCorpora/1/ragFiles:upload",
					validateMultipartForm: func(t *testing.T, r *http.Request) {
						err := r.ParseMultipartForm(32 << 20)
						assert.NoError(t, err)

						// Validate metadata
						metadata := r.FormValue("metadata")
						var metadataObj UploadFileMetadata
						err = json.Unmarshal([]byte(metadata), &metadataObj)
						assert.NoError(t, err)

						assert.Equal(t, "test.md", metadataObj.RagFile.DisplayName)
						assert.Empty(t, metadataObj.RagFile.Description)
						assert.Nil(t, metadataObj.UploadRagFileConfig)

						// Validate file
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
			name:     "Error: API returns error",
			file:     strings.NewReader("test content"),
			fileName: "test.md",
			options:  []UploadFileOption{},
			setup: func(mt *mockTransport) {
				mt.responses = map[string]mockResponse{
					"POST /upload/v1/projects/test-project/locations/test-location/ragCorpora/1/ragFiles:upload": {
						statusCode: http.StatusInternalServerError,
						body:       map[string]interface{}{},
					},
				}
			},
			errorExpected:   true,
			expectedRagFile: File{},
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
			// Prepare mock transport
			mt := newMockTransport(t, tt.expectedReqs)
			tt.setup(mt)

			// Create test target
			client := NewClient(&http.Client{Transport: mt}, "test-project", "test-location")

			// Execute test
			ragFile, err := client.UploadFile(
				context.Background(),
				1,
				tt.file,
				tt.fileName,
				tt.options...,
			)

			// Assert results
			if tt.errorExpected {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedRagFile, ragFile)
			}

			// Verify requests
			mt.verify(t)
		})
	}
}
