package main

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strconv"

	"github.com/alecthomas/kong"
	"golang.org/x/oauth2/google"

	"docgent-backend/internal/infrastructure/google/vertexai/rag/lib"
)

var CLI struct {
	Upload struct {
		File        string `arg:"" help:"Path to the file to upload" type:"path"`
		CorpusID    string `required:"" help:"ID of the RAG corpus"`
		Description string `help:"Description of the file"`
		ChunkSize   int    `help:"Size of each chunk" default:"1000"`
		Overlap     int    `help:"Overlap size between chunks" default:"100"`
	} `cmd:"" help:"Upload a file to the RAG corpus"`

	Delete struct {
		CorpusID string `required:"" help:"ID of the RAG corpus"`
		FileID   string `required:"" help:"ID of the RAG file to delete"`
	} `cmd:"" help:"Delete a file from the RAG corpus"`

	ProjectID string `required:"" help:"Google Cloud Project ID"`
	Location  string `help:"Google Cloud location" default:"us-central1"`
}

func main() {
	ctx := context.Background()

	kongCtx := kong.Parse(&CLI)

	credentials, err := google.FindDefaultCredentials(ctx)
	if err != nil {
		fmt.Printf("Failed to get Google Cloud credentials: %v\n", err)
		os.Exit(1)
	}

	client := lib.NewClientWithCredentials(credentials, CLI.ProjectID, CLI.Location)

	switch kongCtx.Command() {
	case "upload":
		file, err := os.Open(CLI.Upload.File)
		if err != nil {
			fmt.Printf("Failed to open file: %v\n", err)
			os.Exit(1)
		}
		defer file.Close()

		corpusID, err := strconv.ParseInt(CLI.Upload.CorpusID, 10, 64)
		if err != nil {
			fmt.Printf("Invalid corpus ID format: %v\n", err)
			os.Exit(1)
		}

		fileName := filepath.Base(CLI.Upload.File)

		createdFile, err := client.UploadFile(ctx, corpusID, file, fileName, func(o *lib.UploadFileOptions) {
			if CLI.Upload.Description != "" {
				o.Description = CLI.Upload.Description
			}
			o.ChunkingConfig = lib.ChunkingConfig{
				ChunkSize:    CLI.Upload.ChunkSize,
				ChunkOverlap: CLI.Upload.Overlap,
			}
		})

		if err != nil {
			var httpErr *lib.HTTPError
			if errors.As(err, &httpErr) {
				fmt.Printf("Failed to upload file: %s %s\n", httpErr.Status, httpErr.RawBody)
			} else {
				fmt.Printf("Failed to upload file: %v\n", err)
			}
			os.Exit(1)
		}

		fmt.Printf("Successfully uploaded file '%s'\n", createdFile.Name)

	case "delete":
		corpusID, err := strconv.ParseInt(CLI.Delete.CorpusID, 10, 64)
		if err != nil {
			fmt.Printf("Invalid corpus ID format: %v\n", err)
			os.Exit(1)
		}

		fileID, err := strconv.ParseInt(CLI.Delete.FileID, 10, 64)
		if err != nil {
			fmt.Printf("Invalid file ID format: %v\n", err)
			os.Exit(1)
		}

		err = client.DeleteFile(ctx, corpusID, fileID)
		if err != nil {
			var httpErr *lib.HTTPError
			if errors.As(err, &httpErr) {
				fmt.Printf("Failed to delete file: %s %s\n", httpErr.Status, httpErr.RawBody)
			} else {
				fmt.Printf("Failed to delete file: %v\n", err)
			}
			os.Exit(1)
		}

		fmt.Printf("Successfully deleted file from corpus %d\n", corpusID)
	}
}
