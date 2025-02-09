package cli

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strconv"

	"docgent-backend/internal/infrastructure/google/vertexai/rag/lib"
)

func HandleFileUpload(ctx context.Context, cli *CLI, client *lib.Client) error {
	file, err := os.Open(cli.File.Upload.File)
	if err != nil {
		return fmt.Errorf("failed to open file: %v", err)
	}
	defer file.Close()

	corpusID, err := strconv.ParseInt(cli.File.Upload.CorpusID, 10, 64)
	if err != nil {
		return fmt.Errorf("invalid corpus ID format: %v", err)
	}

	fileName := filepath.Base(cli.File.Upload.File)

	createdFile, err := client.UploadFile(ctx, corpusID, file, fileName, func(o *lib.UploadFileOptions) {
		if cli.File.Upload.Description != "" {
			o.Description = cli.File.Upload.Description
		}
		o.ChunkingConfig = lib.ChunkingConfig{
			ChunkSize:    cli.File.Upload.ChunkSize,
			ChunkOverlap: cli.File.Upload.Overlap,
		}
	})

	if err != nil {
		var httpErr *lib.HTTPError
		if errors.As(err, &httpErr) {
			return fmt.Errorf("failed to upload file: %s %s", httpErr.Status, httpErr.RawBody)
		}
		return fmt.Errorf("failed to upload file: %v", err)
	}

	fmt.Printf("Successfully uploaded file '%s'\n", createdFile.Name)
	return nil
}

func HandleFileDelete(ctx context.Context, cli *CLI, client *lib.Client) error {
	corpusID, err := strconv.ParseInt(cli.File.Delete.CorpusID, 10, 64)
	if err != nil {
		return fmt.Errorf("invalid corpus ID format: %v", err)
	}

	fileID, err := strconv.ParseInt(cli.File.Delete.FileID, 10, 64)
	if err != nil {
		return fmt.Errorf("invalid file ID format: %v", err)
	}

	err = client.DeleteFile(ctx, corpusID, fileID)
	if err != nil {
		var httpErr *lib.HTTPError
		if errors.As(err, &httpErr) {
			return fmt.Errorf("failed to delete file: %s %s", httpErr.Status, httpErr.RawBody)
		}
		return fmt.Errorf("failed to delete file: %v", err)
	}

	fmt.Printf("Successfully deleted file from corpus %d\n", corpusID)
	return nil
}
