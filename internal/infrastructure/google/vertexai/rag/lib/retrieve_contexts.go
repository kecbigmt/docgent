package lib

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

type RetrieveContextsRequest struct {
	VertexRagStore VertexRagStore `json:"vertex_rag_store"`
	Query          Query          `json:"query"`
}

type VertexRagStore struct {
	RagResources            RagResources `json:"rag_resources"`
	VectorDistanceThreshold float64      `json:"vector_distance_threshold"`
}

type RagResources struct {
	RagCorpus string `json:"rag_corpus"`
}

type Query struct {
	Text               string             `json:"text"`
	RagRetrievalConfig RagRetrievalConfig `json:"rag_retrieval_config,omitempty"`
}

type RagRetrievalConfig struct {
	TopK int32 `json:"top_k"`
}

type RetrieveContextsResponse struct {
	Contexts RetrieveContexts `json:"contexts"`
}

type RetrieveContexts struct {
	Contexts []*RetrievalContext `json:"contexts"`
}

type RetrievalContext struct {
	SourceURI         string  `json:"sourceUri"`
	SourceDisplayName string  `json:"sourceDisplayName"`
	Text              string  `json:"text"`
	Score             float64 `json:"score"`
}

// RetrieveContexts retrieves contexts from a RAG corpus.
// References:
// - Example: https://cloud.google.com/vertex-ai/generative-ai/docs/model-reference/rag-api-v1#retrieval-query-api
// - Parameters: https://cloud.google.com/vertex-ai/generative-ai/docs/model-reference/rag-api-v1#retrieval-and-prediction-params-api
func (c *Client) RetrieveContexts(ctx context.Context, corpusId int64, query string, similarityTopK int32, vectorDistanceThreshold float64) (RetrieveContextsResponse, error) {
	url := fmt.Sprintf("https://%s-aiplatform.googleapis.com/v1/projects/%s/locations/%s:retrieveContexts", c.location, c.projectID, c.location)
	corpus := fmt.Sprintf("projects/%s/locations/%s/ragCorpora/%d", c.projectID, c.location, corpusId)

	reqBody := RetrieveContextsRequest{
		VertexRagStore: VertexRagStore{
			RagResources: RagResources{
				RagCorpus: corpus,
			},
			VectorDistanceThreshold: vectorDistanceThreshold,
		},
		Query: Query{
			Text: query,
			RagRetrievalConfig: RagRetrievalConfig{
				TopK: similarityTopK,
			},
		},
	}
	reqBodyBytes, err := json.Marshal(reqBody)
	if err != nil {
		return RetrieveContextsResponse{}, err
	}

	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(reqBodyBytes))
	if err != nil {
		return RetrieveContextsResponse{}, err
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return RetrieveContextsResponse{}, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		resBody, _ := io.ReadAll(resp.Body)
		return RetrieveContextsResponse{}, &HTTPError{
			StatusCode: resp.StatusCode,
			Status:     http.StatusText(resp.StatusCode),
			RawBody:    string(resBody),
		}
	}

	var responseBody RetrieveContextsResponse
	if err := json.NewDecoder(resp.Body).Decode(&responseBody); err != nil {
		return RetrieveContextsResponse{}, err
	}

	return responseBody, nil
}
