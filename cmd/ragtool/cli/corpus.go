package cli

import (
	"context"
	"errors"
	"fmt"

	"docgent-backend/internal/infrastructure/google/vertexai/rag/lib"
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
