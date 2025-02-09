package rag

import (
	"context"
	"docgent-backend/internal/application/port"
	"docgent-backend/internal/infrastructure/google/vertexai/rag/lib"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCorpus_Query(t *testing.T) {
	tests := []struct {
		name                    string
		query                   string
		similarityTopK          int32
		vectorDistanceThreshold float64
		setup                   func(*mockTransport)
		expectedDocuments       []port.RAGDocument
		errorExpected           bool
		expectedReqs            []mockRequest
	}{
		{
			name:                    "正常系: 検索結果が返ってくる場合",
			query:                   "test query",
			similarityTopK:          3,
			vectorDistanceThreshold: 0.8,
			setup: func(mt *mockTransport) {
				mt.responses = map[string]mockResponse{
					"POST /v1/projects/test-project/locations/test-location:retrieveContexts": {
						statusCode: http.StatusOK,
						body: lib.RetrieveContextsResponse{
							Contexts: lib.RetrieveContexts{
								Contexts: []*lib.RetrievalContext{
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
			expectedDocuments: []port.RAGDocument{
				{
					Content: "test content 1",
					Source:  "source1.md",
					Score:   0.9,
				},
				{
					Content: "test content 2",
					Source:  "source2.md",
					Score:   0.8,
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
			name:                    "異常系: APIエラーの場合",
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
			expectedDocuments: nil,
			errorExpected:     true,
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
			name:                    "正常系: 検索結果が0件の場合",
			query:                   "test query",
			similarityTopK:          3,
			vectorDistanceThreshold: 0.8,
			setup: func(mt *mockTransport) {
				mt.responses = map[string]mockResponse{
					"POST /v1/projects/test-project/locations/test-location:retrieveContexts": {
						statusCode: http.StatusOK,
						body: lib.RetrieveContextsResponse{
							Contexts: lib.RetrieveContexts{
								Contexts: []*lib.RetrievalContext{},
							},
						},
					},
				}
			},
			expectedDocuments: []port.RAGDocument{},
			errorExpected:     false,
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
			// モックトランスポーターの準備
			mt := newMockTransport(t, tt.expectedReqs)
			tt.setup(mt)

			// テスト対象のインスタンスを作成
			client := lib.NewClient(&http.Client{Transport: mt}, "test-project", "test-location")
			corpus := NewCorpus(client, 1)

			// テスト実行
			documents, err := corpus.Query(
				context.Background(),
				tt.query,
				tt.similarityTopK,
				tt.vectorDistanceThreshold,
			)

			// アサーション
			if tt.errorExpected {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedDocuments, documents)
			}

			// リクエストの検証
			mt.verify(t)
		})
	}
}
