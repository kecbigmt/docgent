package cli

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"docgent/internal/infrastructure/github"
	"docgent/internal/infrastructure/google/vertexai/rag/lib"

	gogithub "github.com/google/go-github/v68/github"
	"golang.org/x/oauth2"
)

// HandleCorpusMigrate migrates a RAG corpus to use GitHub permalinks as displayName
func HandleCorpusMigrate(ctx context.Context, cli *CLI, client *lib.Client) error {
	fmt.Printf("Migrating RAG corpus %s to use GitHub permalinks...\n", cli.Corpus.Migrate.CorpusID)

	// Convert corpus ID to int64
	corpusID, err := strconv.ParseInt(cli.Corpus.Migrate.CorpusID, 10, 64)
	if err != nil {
		return fmt.Errorf("invalid corpus ID: %w", err)
	}

	// List all files in the corpus directly using the lib client
	listFilesResult, err := client.ListFiles(ctx, corpusID)
	if err != nil {
		return fmt.Errorf("failed to list files: %w", err)
	}

	if len(listFilesResult.Files) == 0 {
		fmt.Println("No files found in the corpus. Nothing to migrate.")
		return nil
	}

	fmt.Printf("Found %d files in the corpus.\n", len(listFilesResult.Files))

	// Create GitHub client
	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: cli.Corpus.Migrate.GithubToken},
	)
	tc := oauth2.NewClient(ctx, ts)
	githubClient := gogithub.NewClient(tc)

	// Create GitHub file query service
	fileQueryService := github.NewFileQueryService(
		githubClient,
		cli.Corpus.Migrate.Owner,
		cli.Corpus.Migrate.Repo,
		cli.Corpus.Migrate.Branch,
	)

	// Process each file
	for _, file := range listFilesResult.Files {
		// Extract file name from the file.Name path
		// The file.Name format is: projects/{project}/locations/{location}/ragCorpora/{corpus_id}/ragFiles/{file_id}
		parts := strings.Split(file.Name, "/")
		fileID, err := strconv.ParseInt(parts[len(parts)-1], 10, 64)
		if err != nil {
			fmt.Printf("Warning: Failed to parse file ID from %s: %v. Skipping.\n", file.Name, err)
			continue
		}

		fmt.Printf("Processing file: %s (ID: %d)\n", file.DisplayName, fileID)

		// Skip files that don't look like file paths (might already be permalinks)
		if strings.HasPrefix(file.DisplayName, "http") {
			fmt.Printf("Skipping %s as it appears to already be a URL\n", file.DisplayName)
			continue
		}

		// Get the file content from GitHub
		githubFile, err := fileQueryService.FindFile(ctx, file.DisplayName)
		if err != nil {
			fmt.Printf("Warning: Failed to get file %s from GitHub: %v. Skipping.\n", file.DisplayName, err)
			continue
		}

		// Get the GitHub permalink
		uri, err := fileQueryService.GetURI(ctx, file.DisplayName)
		if err != nil {
			fmt.Printf("Warning: Failed to get URI for file %s: %v. Skipping.\n", file.DisplayName, err)
			continue
		}

		fmt.Printf("Generated permalink: %s\n", uri)

		// Delete the old file
		err = client.DeleteFile(ctx, corpusID, fileID)
		if err != nil {
			return fmt.Errorf("failed to delete file %s (ID: %d): %w", file.DisplayName, fileID, err)
		}

		// Upload the file with the new URI as displayName
		reader := strings.NewReader(githubFile.Content)
		_, err = client.UploadFile(ctx, corpusID, reader, uri.String())
		if err != nil {
			return fmt.Errorf("failed to upload file with URI %s: %w", uri, err)
		}

		fmt.Printf("Successfully migrated file %s to use permalink\n", file.DisplayName)
	}

	fmt.Println("Migration completed successfully!")
	return nil
}
