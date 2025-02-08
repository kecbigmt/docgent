package rag

import (
	"bytes"
	"context"
	"docgent-backend/internal/domain"
	"docgent-backend/internal/infrastructure/google/vertexai/rag/lib"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
)

type mockTransport struct {
	responses    map[string]mockResponse
	requests     []mockRequest
	expectedReqs []mockRequest
}

type mockResponse struct {
	statusCode int
	body       interface{}
}

type mockRequest struct {
	method string
	path   string
	body   interface{}
}

func (m *mockTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	key := req.Method + " " + req.URL.Path

	// リクエストボディの読み取り
	var reqBody interface{}
	if req.Body != nil {
		body, err := io.ReadAll(req.Body)
		if err != nil {
			return nil, err
		}
		// ボディを再度設定（ReadAllで消費されるため）
		req.Body = io.NopCloser(bytes.NewReader(body))

		// JSONデコード
		if len(body) > 0 {
			var v interface{}
			if err := json.Unmarshal(body, &v); err != nil {
				return nil, err
			}
			reqBody = v
		}
	}

	// リクエストを記録
	m.requests = append(m.requests, mockRequest{
		method: req.Method,
		path:   req.URL.Path,
		body:   reqBody,
	})

	if resp, ok := m.responses[key]; ok {
		body, err := json.Marshal(resp.body)
		if err != nil {
			return nil, err
		}
		return &http.Response{
			StatusCode: resp.statusCode,
			Body:       io.NopCloser(bytes.NewReader(body)),
		}, nil
	}
	return &http.Response{
		StatusCode: http.StatusNotFound,
		Body:       io.NopCloser(bytes.NewReader([]byte{})),
	}, nil
}

func (m *mockTransport) verify(t *testing.T) {
	assert.Equal(t, len(m.expectedReqs), len(m.requests), "リクエスト数が一致しません")
	for i, expected := range m.expectedReqs {
		if i >= len(m.requests) {
			t.Errorf("期待されるリクエスト %d が実行されませんでした: %+v", i, expected)
			continue
		}
		actual := m.requests[i]
		assert.Equal(t, expected.method, actual.method, fmt.Sprintf("リクエスト %d のメソッドが一致しません", i))
		assert.Equal(t, expected.path, actual.path, fmt.Sprintf("リクエスト %d のパスが一致しません", i))
		if expected.body != nil {
			assert.Equal(t, expected.body, actual.body, fmt.Sprintf("リクエスト %d のボディが一致しません", i))
		}
	}
}

func TestCorpus_Query(t *testing.T) {
	tests := []struct {
		name                    string
		query                   string
		similarityTopK          int32
		vectorDistanceThreshold float64
		setup                   func(*mockTransport)
		expectedDocuments       []domain.RAGDocument
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
			expectedDocuments: []domain.RAGDocument{
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
								"rag_corpus": "projects/test-project/locations/test-location/ragCorpora/test-corpus",
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
								"rag_corpus": "projects/test-project/locations/test-location/ragCorpora/test-corpus",
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
			expectedDocuments: []domain.RAGDocument{},
			errorExpected:     false,
			expectedReqs: []mockRequest{
				{
					method: "POST",
					path:   "/v1/projects/test-project/locations/test-location:retrieveContexts",
					body: map[string]interface{}{
						"vertex_rag_store": map[string]interface{}{
							"rag_resources": map[string]interface{}{
								"rag_corpus": "projects/test-project/locations/test-location/ragCorpora/test-corpus",
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
			mt := &mockTransport{
				responses:    make(map[string]mockResponse),
				expectedReqs: tt.expectedReqs,
			}
			tt.setup(mt)

			// テスト対象のインスタンスを作成
			client := lib.NewClient(&http.Client{Transport: mt}, "test-project", "test-location")
			corpus := NewCorpus(client, "test-corpus")

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
