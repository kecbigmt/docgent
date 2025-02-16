package cli

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"

	"docgent/internal/infrastructure/google/vertexai/rag/lib"
)

func HandleCorpusCreate(ctx context.Context, cli *CLI, client *lib.Client) error {
	var options []lib.CreateCorpusOption
	if cli.Corpus.Create.Description != "" {
		options = append(options, lib.WithCreateCorpusDescription(cli.Corpus.Create.Description))
	}

	switch cli.Corpus.Create.VectorDB {
	case "pinecone":
		if cli.Corpus.Create.PineconeIndexName == "" || cli.Corpus.Create.PineconeAPIKeySecretVersion == "" {
			return fmt.Errorf("pinecone index name and API key secret version are required")
		}
		config := lib.PineconeConfig{
			IndexName:    cli.Corpus.Create.PineconeIndexName,
			APIKeyConfig: lib.APIKeyConfig{APIKeySecretVersion: cli.Corpus.Create.PineconeAPIKeySecretVersion},
		}
		config.RAGEmbeddingModelConfig = maybeSetEmbeddingConfig(cli.Corpus.Create.EmbeddingPredictionEndpoint)
		options = append(options, lib.WithVectorDBConfig(config))
	case "vertex_vector_search":
		if cli.Corpus.Create.VectorSearchIndex == "" || cli.Corpus.Create.VectorSearchIndexEndpoint == "" {
			return fmt.Errorf("vertex vector search index and endpoint are required")
		}
		config := lib.VertexVectorSearchConfig{
			Index:         cli.Corpus.Create.VectorSearchIndex,
			IndexEndpoint: cli.Corpus.Create.VectorSearchIndexEndpoint,
		}
		config.RAGEmbeddingModelConfig = maybeSetEmbeddingConfig(cli.Corpus.Create.EmbeddingPredictionEndpoint)
		options = append(options, lib.WithVectorDBConfig(config))
	case "rag_managed_db":
		config := lib.RAGManagedDBConfig{
			RAGEmbeddingModelConfig: maybeSetEmbeddingConfig(cli.Corpus.Create.EmbeddingPredictionEndpoint),
		}
		options = append(options, lib.WithVectorDBConfig(config))
	default:
		return fmt.Errorf("invalid vector database type: %s", cli.Corpus.Create.VectorDB)
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

func HandleCorpusDelete(ctx context.Context, cli *CLI, client *lib.Client) error {
	corpusID, err := strconv.ParseInt(cli.Corpus.Delete.CorpusID, 10, 64)
	if err != nil {
		return fmt.Errorf("invalid corpus ID format: %v", err)
	}

	err = client.DeleteCorpus(ctx, corpusID)
	if err != nil {
		var httpErr *lib.HTTPError
		if errors.As(err, &httpErr) {
			return fmt.Errorf("failed to delete corpus: %s %s", httpErr.Status, httpErr.RawBody)
		}
		return fmt.Errorf("failed to delete corpus: %v", err)
	}

	fmt.Printf("Successfully deleted corpus %d\n", corpusID)
	return nil
}

func maybeSetEmbeddingConfig(endpoint string) *lib.RAGEmbeddingModelConfig {
	if endpoint == "" {
		return nil
	}
	return &lib.RAGEmbeddingModelConfig{
		VertexPredictionEndpoint: lib.VertexPredictionEndpoint{Endpoint: endpoint},
	}
}
