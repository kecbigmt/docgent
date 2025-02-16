package lib

import (
	"context"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestClient_CreateCorpus(t *testing.T) {
	tests := []struct {
		name          string
		displayName   string
		options       []CreateCorpusOption
		setup         func(*mockTransport)
		errorExpected bool
		expectedReqs  []mockRequest
	}{
		{
			name:        "Success: Create corpus with display name only",
			displayName: "test-corpus",
			options:     []CreateCorpusOption{},
			setup: func(mt *mockTransport) {
				mt.responses = map[string]mockResponse{
					"POST /v1/projects/test-project/locations/test-location/ragCorpora": {
						statusCode: http.StatusOK,
						body:       map[string]interface{}{},
					},
				}
			},
			errorExpected: false,
			expectedReqs: []mockRequest{
				{
					method: "POST",
					path:   "/v1/projects/test-project/locations/test-location/ragCorpora",
					body: map[string]interface{}{
						"display_name": "test-corpus",
					},
				},
			},
		},
		{
			name:        "Success: Create corpus with description",
			displayName: "test-corpus",
			options: []CreateCorpusOption{
				WithCreateCorpusDescription("test description"),
			},
			setup: func(mt *mockTransport) {
				mt.responses = map[string]mockResponse{
					"POST /v1/projects/test-project/locations/test-location/ragCorpora": {
						statusCode: http.StatusOK,
						body:       map[string]interface{}{},
					},
				}
			},
			errorExpected: false,
			expectedReqs: []mockRequest{
				{
					method: "POST",
					path:   "/v1/projects/test-project/locations/test-location/ragCorpora",
					body: map[string]interface{}{
						"display_name": "test-corpus",
						"description":  "test description",
					},
				},
			},
		},
		{
			name:        "Error: API returns error",
			displayName: "test-corpus",
			options:     []CreateCorpusOption{},
			setup: func(mt *mockTransport) {
				mt.responses = map[string]mockResponse{
					"POST /v1/projects/test-project/locations/test-location/ragCorpora": {
						statusCode: http.StatusInternalServerError,
						body:       map[string]interface{}{},
					},
				}
			},
			errorExpected: true,
			expectedReqs: []mockRequest{
				{
					method: "POST",
					path:   "/v1/projects/test-project/locations/test-location/ragCorpora",
					body: map[string]interface{}{
						"display_name": "test-corpus",
					},
				},
			},
		},
		{
			name:        "Success: Create corpus with RAGManagedDBConfig",
			displayName: "test-corpus",
			options: []CreateCorpusOption{
				WithVectorDBConfig(RAGManagedDBConfig{
					RAGEmbeddingModelConfig: &RAGEmbeddingModelConfig{
						VertexPredictionEndpoint: VertexPredictionEndpoint{
							Endpoint: "test-endpoint",
						},
					},
				}),
			},
			setup: func(mt *mockTransport) {
				mt.responses = map[string]mockResponse{
					"POST /v1/projects/test-project/locations/test-location/ragCorpora": {
						statusCode: http.StatusOK,
						body:       map[string]interface{}{},
					},
				}
			},
			errorExpected: false,
			expectedReqs: []mockRequest{
				{
					method: "POST",
					path:   "/v1/projects/test-project/locations/test-location/ragCorpora",
					body: map[string]interface{}{
						"display_name": "test-corpus",
						"vector_db_config": map[string]interface{}{
							"rag_embedding_model_config": map[string]interface{}{
								"vertex_prediction_endpoint": map[string]interface{}{
									"endpoint": "test-endpoint",
								},
							},
						},
					},
				},
			},
		},
		{
			name:        "Success: Create corpus with PineconeConfig",
			displayName: "test-corpus",
			options: []CreateCorpusOption{
				WithVectorDBConfig(PineconeConfig{
					IndexName: "test-index",
					APIKeyConfig: APIKeyConfig{
						APIKeySecretVersion: "test-secret-version",
					},
					RAGEmbeddingModelConfig: &RAGEmbeddingModelConfig{
						VertexPredictionEndpoint: VertexPredictionEndpoint{
							Endpoint: "test-endpoint",
						},
					},
				}),
			},
			setup: func(mt *mockTransport) {
				mt.responses = map[string]mockResponse{
					"POST /v1/projects/test-project/locations/test-location/ragCorpora": {
						statusCode: http.StatusOK,
						body:       map[string]interface{}{},
					},
				}
			},
			errorExpected: false,
			expectedReqs: []mockRequest{
				{
					method: "POST",
					path:   "/v1/projects/test-project/locations/test-location/ragCorpora",
					body: map[string]interface{}{
						"display_name": "test-corpus",
						"vector_db_config": map[string]interface{}{
							"pinecone": map[string]interface{}{
								"index_name": "test-index",
							},
							"api_auth": map[string]interface{}{
								"api_key_config": map[string]interface{}{
									"api_key_secret_version": "test-secret-version",
								},
							},
							"rag_embedding_model_config": map[string]interface{}{
								"vertex_prediction_endpoint": map[string]interface{}{
									"endpoint": "test-endpoint",
								},
							},
						},
					},
				},
			},
		},
		{
			name:        "Success: Create corpus with VertexVectorSearchConfig",
			displayName: "test-corpus",
			options: []CreateCorpusOption{
				WithVectorDBConfig(VertexVectorSearchConfig{
					Index:         "test-index",
					IndexEndpoint: "test-index-endpoint",
					RAGEmbeddingModelConfig: &RAGEmbeddingModelConfig{
						VertexPredictionEndpoint: VertexPredictionEndpoint{
							Endpoint: "test-endpoint",
						},
					},
				}),
			},
			setup: func(mt *mockTransport) {
				mt.responses = map[string]mockResponse{
					"POST /v1/projects/test-project/locations/test-location/ragCorpora": {
						statusCode: http.StatusOK,
						body:       map[string]interface{}{},
					},
				}
			},
			errorExpected: false,
			expectedReqs: []mockRequest{
				{
					method: "POST",
					path:   "/v1/projects/test-project/locations/test-location/ragCorpora",
					body: map[string]interface{}{
						"display_name": "test-corpus",
						"vector_db_config": map[string]interface{}{
							"vertex_vector_search": map[string]interface{}{
								"index":          "test-index",
								"index_endpoint": "test-index-endpoint",
							},
							"rag_embedding_model_config": map[string]interface{}{
								"vertex_prediction_endpoint": map[string]interface{}{
									"endpoint": "test-endpoint",
								},
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
			err := client.CreateCorpus(
				context.Background(),
				tt.displayName,
				tt.options...,
			)

			// Assert results
			if tt.errorExpected {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

			// Verify requests
			mt.verify(t)
		})
	}
}
