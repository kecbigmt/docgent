package lib

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

// CreateCorpus creates a new RAG corpus.
// References:
// - Example: https://cloud.google.com/vertex-ai/generative-ai/docs/model-reference/rag-api-v1#create-a-rag-corpus-example-api
// - Parameters: https://cloud.google.com/vertex-ai/generative-ai/docs/model-reference/rag-api-v1#corpus-management-params-api
func (c *Client) CreateCorpus(ctx context.Context, displayName string, options ...CreateCorpusOption) error {
	params := CreateCorpusParams{}
	for _, option := range options {
		option(&params)
	}

	params.DisplayName = displayName

	reqBodyBytes, err := json.Marshal(params)
	if err != nil {
		return err
	}

	url := fmt.Sprintf("https://%s-aiplatform.googleapis.com/v1/projects/%s/locations/%s/ragCorpora", c.location, c.projectID, c.location)
	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(reqBodyBytes))
	if err != nil {
		return err
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return err
	}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("failed to create corpus: %s %s", resp.Status, string(body))
	}

	return nil
}

type CreateCorpusOption func(*CreateCorpusParams)

type CreateCorpusParams struct {
	DisplayName    string                 `json:"display_name"`
	Description    string                 `json:"description,omitempty"`
	VectorDBConfig map[string]interface{} `json:"vector_db_config,omitempty"`
}

func WithCreateCorpusDescription(description string) CreateCorpusOption {
	return func(p *CreateCorpusParams) {
		p.Description = description
	}
}

// WithVectorDBConfig sets the vector database configuration for the corpus.
// Supported vector databases:
// - RAG Managed DB
// - Pinecone
// - Vertex Vector Search
// Reference: https://cloud.google.com/vertex-ai/generative-ai/docs/model-reference/rag-api-v1
func WithVectorDBConfig(config VectorDBConfig) CreateCorpusOption {
	return func(p *CreateCorpusParams) {
		p.VectorDBConfig = map[string]interface{}{}
		var ragEmbedding *RAGEmbeddingModelConfig

		config.Match(VectorDBConfigCases{
			RAGManagedDBConfig: func(c RAGManagedDBConfig) {
				ragEmbedding = c.RAGEmbeddingModelConfig
			},
			PineconeConfig: func(c PineconeConfig) {
				p.VectorDBConfig["pinecone"] = map[string]interface{}{
					"index_name": c.IndexName,
				}
				p.VectorDBConfig["api_auth"] = map[string]interface{}{
					"api_key_config": map[string]interface{}{
						"api_key_secret_version": c.APIKeyConfig.APIKeySecretVersion,
					},
				}
				ragEmbedding = c.RAGEmbeddingModelConfig
			},
			VertexVectorSearchConfig: func(c VertexVectorSearchConfig) {
				p.VectorDBConfig["vertex_vector_search"] = map[string]interface{}{
					"index":          c.Index,
					"index_endpoint": c.IndexEndpoint,
				}
				ragEmbedding = c.RAGEmbeddingModelConfig
			},
		})
		if ragEmbedding != nil {
			p.VectorDBConfig["rag_embedding_model_config"] = buildRAGEmbeddingModelConfig(ragEmbedding)
		}
	}
}

func buildRAGEmbeddingModelConfig(emb *RAGEmbeddingModelConfig) map[string]interface{} {
	if emb == nil {
		return nil
	}
	return map[string]interface{}{
		"vertex_prediction_endpoint": map[string]interface{}{
			"endpoint": emb.VertexPredictionEndpoint.Endpoint,
		},
	}
}

// VectorDBConfig is the interface for the vector database configuration.
type VectorDBConfig interface {
	Match(cs VectorDBConfigCases)
}

// VectorDBConfigCases is the interface for the vector database configuration cases.
type VectorDBConfigCases struct {
	RAGManagedDBConfig       func(RAGManagedDBConfig)
	PineconeConfig           func(PineconeConfig)
	VertexVectorSearchConfig func(VertexVectorSearchConfig)
}

type RAGManagedDBConfig struct {
	RAGEmbeddingModelConfig *RAGEmbeddingModelConfig
}

func (c RAGManagedDBConfig) Match(cs VectorDBConfigCases) {
	cs.RAGManagedDBConfig(c)
}

type PineconeConfig struct {
	IndexName               string
	APIKeyConfig            APIKeyConfig
	RAGEmbeddingModelConfig *RAGEmbeddingModelConfig
}

type APIKeyConfig struct {
	APIKeySecretVersion string
}

func (c PineconeConfig) Match(cs VectorDBConfigCases) {
	cs.PineconeConfig(c)
}

type VertexVectorSearchConfig struct {
	Index                   string
	IndexEndpoint           string
	RAGEmbeddingModelConfig *RAGEmbeddingModelConfig
}

func (c VertexVectorSearchConfig) Match(cs VectorDBConfigCases) {
	cs.VertexVectorSearchConfig(c)
}

type RAGEmbeddingModelConfig struct {
	VertexPredictionEndpoint VertexPredictionEndpoint
}

type VertexPredictionEndpoint struct {
	Endpoint string
}
