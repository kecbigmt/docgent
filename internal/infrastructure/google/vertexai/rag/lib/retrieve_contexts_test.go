package lib

import (
	"context"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestClient_RetrieveContexts(t *testing.T) {
	tests := []struct {
		name                    string
		query                   string
		similarityTopK          int32
		vectorDistanceThreshold float64
		setup                   func(*mockTransport)
		expectedResponse        RetrieveContextsResponse
		errorExpected           bool
		expectedReqs            []mockRequest
	}{
		{
			name:                    "Success: Returns search results",
			query:                   "test query",
			similarityTopK:          3,
			vectorDistanceThreshold: 0.8,
			setup: func(mt *mockTransport) {
				mt.responses = map[string]mockResponse{
					"POST /v1/projects/test-project/locations/test-location:retrieveContexts": {
						statusCode: http.StatusOK,
						body: RetrieveContextsResponse{
							Contexts: RetrieveContexts{
								Contexts: []*RetrievalContext{
									{
										Text:      "test content 1",
										SourceUri: "source1.md",
										Score:     0.9,
									},
									{
										Text:      "test content 2",
										SourceUri: "source2.md",
										Score:     0.8,
									},
								},
							},
						},
					},
				}
			},
			expectedResponse: RetrieveContextsResponse{
				Contexts: RetrieveContexts{
					Contexts: []*RetrievalContext{
						{
							Text:      "test content 1",
							SourceUri: "source1.md",
							Score:     0.9,
						},
						{
							Text:      "test content 2",
							SourceUri: "source2.md",
							Score:     0.8,
						},
					},
				},
			},
			errorExpected: false,
			expectedReqs: []mockRequest{
				{
					method: "POST",
					path:   "/v1/projects/test-project/locations/test-location:retrieveContexts",
					body: map[string]interface{}{
						"vertex_rag_store": map[string]interface{}{
							"rag_resources": map[string]interface{}{
								"rag_corpus": "projects/test-project/locations/test-location/ragCorpora/1",
							},
							"vector_distance_threshold": 0.8,
						},
						"query": map[string]interface{}{
							"text": "test query",
							"rag_retrieval_config": map[string]interface{}{
								"top_k": float64(3),
							},
						},
					},
				},
			},
		},
		{
			name:                    "Error: API returns error",
			query:                   "test query",
			similarityTopK:          3,
			vectorDistanceThreshold: 0.8,
			setup: func(mt *mockTransport) {
				mt.responses = map[string]mockResponse{
					"POST /v1/projects/test-project/locations/test-location:retrieveContexts": {
						statusCode: http.StatusInternalServerError,
						body:       map[string]interface{}{},
					},
				}
			},
			expectedResponse: RetrieveContextsResponse{},
			errorExpected:    true,
			expectedReqs: []mockRequest{
				{
					method: "POST",
					path:   "/v1/projects/test-project/locations/test-location:retrieveContexts",
					body: map[string]interface{}{
						"vertex_rag_store": map[string]interface{}{
							"rag_resources": map[string]interface{}{
								"rag_corpus": "projects/test-project/locations/test-location/ragCorpora/1",
							},
							"vector_distance_threshold": 0.8,
						},
						"query": map[string]interface{}{
							"text": "test query",
							"rag_retrieval_config": map[string]interface{}{
								"top_k": float64(3),
							},
						},
					},
				},
			},
		},
		{
			name:                    "Success: Returns empty results",
			query:                   "test query",
			similarityTopK:          3,
			vectorDistanceThreshold: 0.8,
			setup: func(mt *mockTransport) {
				mt.responses = map[string]mockResponse{
					"POST /v1/projects/test-project/locations/test-location:retrieveContexts": {
						statusCode: http.StatusOK,
						body: RetrieveContextsResponse{
							Contexts: RetrieveContexts{
								Contexts: []*RetrievalContext{},
							},
						},
					},
				}
			},
			expectedResponse: RetrieveContextsResponse{
				Contexts: RetrieveContexts{
					Contexts: []*RetrievalContext{},
				},
			},
			errorExpected: false,
			expectedReqs: []mockRequest{
				{
					method: "POST",
					path:   "/v1/projects/test-project/locations/test-location:retrieveContexts",
					body: map[string]interface{}{
						"vertex_rag_store": map[string]interface{}{
							"rag_resources": map[string]interface{}{
								"rag_corpus": "projects/test-project/locations/test-location/ragCorpora/1",
							},
							"vector_distance_threshold": 0.8,
						},
						"query": map[string]interface{}{
							"text": "test query",
							"rag_retrieval_config": map[string]interface{}{
								"top_k": float64(3),
							},
						},
					},
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Prepare mock transport
			mt := newMockTransport(t, tt.expectedReqs)
			tt.setup(mt)

			// Create test target
			client := NewClient(&http.Client{Transport: mt}, "test-project", "test-location")

			// Execute test
			response, err := client.RetrieveContexts(
				context.Background(),
				1,
				tt.query,
				tt.similarityTopK,
				tt.vectorDistanceThreshold,
			)

			// Assert results
			if tt.errorExpected {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedResponse, response)
			}

			// Verify requests
			mt.verify(t)
		})
	}
}
