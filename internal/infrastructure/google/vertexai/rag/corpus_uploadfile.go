package rag

import (
	"context"
	"io"

	"docgent/internal/application/port"
	"docgent/internal/domain/data"
	"docgent/internal/infrastructure/google/vertexai/rag/lib"
)

func (c *Corpus) UploadFile(ctx context.Context, file io.Reader, uri *data.URI, options ...port.RAGCorpusUploadFileOption) error {
	uploadFileOptions := &port.RAGCorpusUploadFileOptions{}
	for _, option := range options {
		option(uploadFileOptions)
	}

	_, err := c.client.UploadFile(ctx, c.corpusId, file, uri.String(), func(o *lib.UploadFileOptions) {
		if uploadFileOptions.Description != "" {
			o.Description = uploadFileOptions.Description
		}
		if uploadFileOptions.ChunkingConfig != (port.ChunkingConfig{}) {
			o.ChunkingConfig = lib.ChunkingConfig{
				ChunkSize:    uploadFileOptions.ChunkingConfig.ChunkSize,
				ChunkOverlap: uploadFileOptions.ChunkingConfig.ChunkOverlap,
			}
		}
	})

	return err
}
