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
	Text           string `json:"text"`
	SimilarityTopK int32  `json:"similarity_top_k"`
}

type RetrieveContextsResponse struct {
	Contexts []*RetrievalContext `json:"contexts"`
}

type RetrievalContext struct {
	SourceUri         string  `json:"source_uri"`
	SourceDisplayName string  `json:"source_display_name"`
	Text              string  `json:"text"`
	Score             float64 `json:"score"`
}

func (c *Client) RetrieveContexts(ctx context.Context, corpusName string, query string, similarityTopK int32, vectorDistanceThreshold float64) (RetrieveContextsResponse, error) {
	url := fmt.Sprintf("https://%s-aiplatform.googleapis.com/v1/projects/%s/locations/%s:retrieveContexts", c.location, c.projectID, c.location)

	reqBody := RetrieveContextsRequest{
		VertexRagStore: VertexRagStore{
			RagResources: RagResources{
				RagCorpus: corpusName,
			},
			VectorDistanceThreshold: vectorDistanceThreshold,
		},
		Query: Query{
			Text:           query,
			SimilarityTopK: similarityTopK,
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
		return RetrieveContextsResponse{}, &HTTPError{
			StatusCode: resp.StatusCode,
			Status:     http.StatusText(resp.StatusCode),
		}
	}

	resBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return RetrieveContextsResponse{}, err
	}

	var responseBody RetrieveContextsResponse
	err = json.Unmarshal(resBody, &responseBody)
	if err != nil {
		return RetrieveContextsResponse{}, err
	}

	return responseBody, nil
}
