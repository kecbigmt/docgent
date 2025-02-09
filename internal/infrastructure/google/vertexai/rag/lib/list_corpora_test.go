package lib

import (
	"context"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestClient_ListCorpora(t *testing.T) {
	tests := []struct {
		name           string
		options        []ListCorpusOption
		setup          func(*mockTransport)
		expectedResult ListCorporaResult
		errorExpected  bool
		expectedReqs   []mockRequest
	}{
		{
			name:    "正常系: オプションなしでコーパスの一覧を取得できる",
			options: []ListCorpusOption{},
			setup: func(mt *mockTransport) {
				mt.responses = map[string]mockResponse{
					"GET /v1/projects/test-project/locations/test-location/ragCorpora": {
						statusCode: http.StatusOK,
						body: map[string]interface{}{
							"ragCorpora": []map[string]interface{}{
								{
									"name":        "projects/test-project/locations/test-location/ragCorpora/test-corpus",
									"displayName": "Test Corpus",
									"description": "Test Description",
									"corpusStatus": map[string]interface{}{
										"state": "ACTIVE",
									},
								},
							},
						},
					},
				}
			},
			expectedResult: ListCorporaResult{
				Corpora: []Corpus{
					{
						Name:        "projects/test-project/locations/test-location/ragCorpora/test-corpus",
						DisplayName: "Test Corpus",
						Description: "Test Description",
						CorpusStatus: CorpusStatus{
							State: CorpusStateActive,
						},
					},
				},
			},
			errorExpected: false,
			expectedReqs: []mockRequest{
				{
					method: "GET",
					path:   "/v1/projects/test-project/locations/test-location/ragCorpora",
				},
			},
		},
		{
			name: "正常系: ページングオプションを指定してコーパスの一覧を取得できる",
			options: []ListCorpusOption{
				WithListCorpusPageSize(10),
				WithListCorpusPageToken("next-page-token"),
			},
			setup: func(mt *mockTransport) {
				mt.responses = map[string]mockResponse{
					"GET /v1/projects/test-project/locations/test-location/ragCorpora": {
						statusCode: http.StatusOK,
						body: map[string]interface{}{
							"ragCorpora": []map[string]interface{}{
								{
									"name":        "projects/test-project/locations/test-location/ragCorpora/test-corpus-2",
									"displayName": "Test Corpus 2",
									"description": "Test Description 2",
									"corpusStatus": map[string]interface{}{
										"state": "ACTIVE",
									},
								},
							},
							"nextPageToken": "next-next-page-token",
						},
					},
				}
			},
			expectedResult: ListCorporaResult{
				Corpora: []Corpus{
					{
						Name:        "projects/test-project/locations/test-location/ragCorpora/test-corpus-2",
						DisplayName: "Test Corpus 2",
						Description: "Test Description 2",
						CorpusStatus: CorpusStatus{
							State: CorpusStateActive,
						},
					},
				},
				NextPageToken: "next-next-page-token",
			},
			errorExpected: false,
			expectedReqs: []mockRequest{
				{
					method: "GET",
					path:   "/v1/projects/test-project/locations/test-location/ragCorpora",
				},
			},
		},
		{
			name:    "異常系: APIがエラーを返す",
			options: []ListCorpusOption{},
			setup: func(mt *mockTransport) {
				mt.responses = map[string]mockResponse{
					"GET /v1/projects/test-project/locations/test-location/ragCorpora": {
						statusCode: http.StatusInternalServerError,
						body:       map[string]interface{}{},
					},
				}
			},
			expectedResult: ListCorporaResult{},
			errorExpected:  true,
			expectedReqs: []mockRequest{
				{
					method: "GET",
					path:   "/v1/projects/test-project/locations/test-location/ragCorpora",
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// モックトランスポートの準備
			mt := newMockTransport(t, tt.expectedReqs)
			tt.setup(mt)

			// テスト用のクライアントを作成
			client := NewClient(&http.Client{Transport: mt}, "test-project", "test-location")

			// テスト実行
			result, err := client.ListCorpora(context.Background(), tt.options...)

			// 結果の検証
			if tt.errorExpected {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedResult, result)
			}

			// リクエストの検証
			mt.verify(t)
		})
	}
}
