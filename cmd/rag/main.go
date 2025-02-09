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
	File        string `arg:"" help:"Path to the file to upload" type:"path"`
	CorpusID    string `required:"" help:"ID of the RAG corpus"`
	Description string `help:"Description of the file"`
	ChunkSize   int    `help:"Size of each chunk" default:"1000"`
	Overlap     int    `help:"Overlap size between chunks" default:"100"`
	ProjectID   string `required:"" help:"Google Cloud Project ID"`
	Location    string `help:"Google Cloud location" default:"us-central1"`
}

func main() {
	ctx := context.Background()

	kong.Parse(&CLI)

	file, err := os.Open(CLI.File)
	if err != nil {
		fmt.Printf("Failed to open file: %v\n", err)
		os.Exit(1)
	}
	defer file.Close()

	credentials, err := google.FindDefaultCredentials(ctx)
	if err != nil {
		fmt.Printf("Failed to get Google Cloud credentials: %v\n", err)
		os.Exit(1)
	}

	client := lib.NewClientWithCredentials(credentials, CLI.ProjectID, CLI.Location)

	corpusID, err := strconv.ParseInt(CLI.CorpusID, 10, 64)
	if err != nil {
		fmt.Printf("Invalid corpus ID format: %v\n", err)
		os.Exit(1)
	}

	fileName := filepath.Base(CLI.File)

	createdFile, err := client.UploadFile(ctx, corpusID, file, fileName, func(o *lib.UploadFileOptions) {
		if CLI.Description != "" {
			o.Description = CLI.Description
		}
		o.ChunkingConfig = lib.ChunkingConfig{
			ChunkSize:    CLI.ChunkSize,
			ChunkOverlap: CLI.Overlap,
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
}
