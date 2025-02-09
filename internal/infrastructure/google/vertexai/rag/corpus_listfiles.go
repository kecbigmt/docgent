package rag

import (
	"context"
	"docgent-backend/internal/application/port"
	"fmt"
	"strconv"
	"strings"

	"docgent-backend/internal/infrastructure/google/vertexai/rag/lib"
)

func (c *Corpus) ListFiles(ctx context.Context) ([]port.RAGFile, error) {
	var nextPageToken string

	var ragFiles []port.RAGFile
	for {
		options := []lib.ListFilesOption{}
		if nextPageToken != "" {
			options = append(options, lib.WithListFilesPageToken(nextPageToken))
		}

		filesResult, err := c.client.ListFiles(ctx, c.corpusId, options...)
		if err != nil {
			return nil, fmt.Errorf("failed to list files: %w", err)
		}

		for _, file := range filesResult.Files {
			parts := strings.Split(file.Name, "/")
			idStr := parts[len(parts)-1]
			id, err := strconv.ParseInt(idStr, 10, 64)
			if err != nil {
				return nil, fmt.Errorf("failed to parse file ID: %w", err)
			}
			ragFiles = append(ragFiles, port.RAGFile{
				ID:          id,
				FileName:    file.Name,
				Description: file.Description,
			})
		}

		if filesResult.NextPageToken == "" {
			break
		}

		nextPageToken = filesResult.NextPageToken
	}

	return ragFiles, nil
}
