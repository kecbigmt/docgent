package cli

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"docgent/internal/infrastructure/google/vertexai/rag/lib"
)

func HandleCorpusCreate(ctx context.Context, cli *CLI, client *lib.Client) error {
	var options []lib.CreateCorpusOption
	if cli.Corpus.Create.Description != "" {
		options = append(options, lib.WithCreateCorpusDescription(cli.Corpus.Create.Description))
	}

	err := client.CreateCorpus(ctx, cli.Corpus.Create.DisplayName, options...)
	if err != nil {
		var httpErr *lib.HTTPError
		if errors.As(err, &httpErr) {
			return fmt.Errorf("failed to create corpus: %s %s", httpErr.Status, httpErr.RawBody)
		}
		return fmt.Errorf("failed to create corpus: %v", err)
	}

	fmt.Printf("Successfully created corpus '%s'\n", cli.Corpus.Create.DisplayName)
	return nil
}

func HandleCorpusList(ctx context.Context, cli *CLI, client *lib.Client) error {
	var options []lib.ListCorpusOption
	options = append(options, lib.WithListCorpusPageSize(cli.Corpus.List.PageSize))
	if cli.Corpus.List.PageToken != "" {
		options = append(options, lib.WithListCorpusPageToken(cli.Corpus.List.PageToken))
	}

	result, err := client.ListCorpora(ctx, options...)
	if err != nil {
		return fmt.Errorf("failed to list corpora: %w", err)
	}

	output, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal response: %w", err)
	}

	fmt.Println(string(output))
	return nil
}
