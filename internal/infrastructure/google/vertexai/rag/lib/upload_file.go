package lib

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
)

type UploadFileMetadata struct {
	RagFile             *UploadFileMetadataRagFile `json:"rag_file"`
	UploadRagFileConfig *UploadRagFileConfig       `json:"upload_rag_file_config"`
}

type UploadFileMetadataRagFile struct {
	DisplayName string `json:"display_name"`
	Description string `json:"description"`
}

type UploadRagFileConfig struct {
	RagFileTransformationConfig *RagFileTransformationConfig `json:"rag_file_transformation_config"`
}

type RagFileTransformationConfig struct {
	RagFileChunkingConfig *RagFileChunkingConfig `json:"rag_file_chunking_config"`
}

type RagFileChunkingConfig struct {
	FixedLengthChunking *FixedLengthChunking `json:"fixed_length_chunking"`
}

type FixedLengthChunking struct {
	ChunkSize    int32 `json:"chunk_size"`
	ChunkOverlap int32 `json:"chunk_overlap"`
}

type UploadFileResponse struct {
	RagFile File `json:"ragFile"`
}

// UploadFile uploads a file to a RAG corpus.
// References:
// - Example: https://cloud.google.com/vertex-ai/generative-ai/docs/model-reference/rag-api-v1#upload-a-rag-file-example-api
// - Parameters: https://cloud.google.com/vertex-ai/generative-ai/docs/model-reference/rag-api-v1#parameters-list
func (c *Client) UploadFile(ctx context.Context, corpusId int64, file io.Reader, fileName string, options ...UploadFileOption) (File, error) {
	uploadFileOptions := &UploadFileOptions{}
	for _, option := range options {
		option(uploadFileOptions)
	}

	metadata := &UploadFileMetadata{
		RagFile: &UploadFileMetadataRagFile{
			DisplayName: fileName,
		},
	}

	if uploadFileOptions.Description != "" {
		metadata.RagFile.Description = uploadFileOptions.Description
	}

	if uploadFileOptions.ChunkingConfig != (ChunkingConfig{}) {
		metadata.UploadRagFileConfig = &UploadRagFileConfig{
			RagFileTransformationConfig: &RagFileTransformationConfig{
				RagFileChunkingConfig: &RagFileChunkingConfig{
					FixedLengthChunking: &FixedLengthChunking{
						ChunkSize:    int32(uploadFileOptions.ChunkingConfig.ChunkSize),
						ChunkOverlap: int32(uploadFileOptions.ChunkingConfig.ChunkOverlap),
					},
				},
			},
		}
	}

	metadataBytes, err := json.Marshal(metadata)
	if err != nil {
		return File{}, err
	}

	var requestBody bytes.Buffer
	writer := multipart.NewWriter(&requestBody)

	if err := writer.WriteField("metadata", string(metadataBytes)); err != nil {
		return File{}, err
	}

	filePart, err := writer.CreateFormFile("file", fileName)
	if err != nil {
		return File{}, err
	}

	if _, err := io.Copy(filePart, file); err != nil {
		return File{}, err
	}

	if err := writer.Close(); err != nil {
		return File{}, err
	}

	url := fmt.Sprintf("https://%s-aiplatform.googleapis.com/upload/v1/projects/%s/locations/%s/ragCorpora/%d/ragFiles:upload", c.location, c.projectID, c.location, corpusId)
	req, err := http.NewRequestWithContext(ctx, "POST", url, &requestBody)
	if err != nil {
		return File{}, err
	}

	req.Header.Set("Content-Type", writer.FormDataContentType())
	req.Header.Set("X-Goog-Upload-Protocol", "multipart")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return File{}, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return File{}, fmt.Errorf("failed to upload file: %w", &HTTPError{
			StatusCode: resp.StatusCode,
			Status:     resp.Status,
			RawBody:    string(body),
		})
	}

	var responseBody UploadFileResponse
	if err := json.NewDecoder(resp.Body).Decode(&responseBody); err != nil {
		return File{}, err
	}

	return responseBody.RagFile, nil
}

type UploadFileOption func(*UploadFileOptions)

type UploadFileOptions struct {
	Description    string
	ChunkingConfig ChunkingConfig
}

func WithUploadFileDescription(description string) UploadFileOption {
	return func(o *UploadFileOptions) {
		o.Description = description
	}
}

func WithUploadFileChunkingConfig(chunkSize int, chunkOverlap int) UploadFileOption {
	return func(o *UploadFileOptions) {
		o.ChunkingConfig = ChunkingConfig{
			ChunkSize:    chunkSize,
			ChunkOverlap: chunkOverlap,
		}
	}
}

type ChunkingConfig struct {
	ChunkSize    int
	ChunkOverlap int
}
