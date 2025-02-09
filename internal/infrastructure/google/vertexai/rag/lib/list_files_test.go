package lib

import (
	"context"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestClient_ListFiles(t *testing.T) {
	tests := []struct {
		name           string
		corpusId       int64
		options        []ListFilesOption
		setup          func(*mockTransport)
		expectedResult ListFilesResult
		errorExpected  bool
		expectedReqs   []mockRequest
	}{
		{
			name:     "Success: List files without options",
			corpusId: 1,
			options:  []ListFilesOption{},
			setup: func(mt *mockTransport) {
				mt.responses = map[string]mockResponse{
					"GET /v1/projects/test-project/locations/test-location/ragCorpora/1/ragFiles": {
						statusCode: http.StatusOK,
						body: map[string]interface{}{
							"ragFiles": []map[string]interface{}{
								{
									"name":        "projects/test-project/locations/test-location/ragCorpora/1/ragFiles/123",
									"displayName": "Test File",
									"description": "Test Description",
								},
							},
						},
					},
				}
			},
			expectedResult: ListFilesResult{
				Files: []File{
					{
						Name:        "projects/test-project/locations/test-location/ragCorpora/1/ragFiles/123",
						DisplayName: "Test File",
						Description: "Test Description",
					},
				},
			},
			errorExpected: false,
			expectedReqs: []mockRequest{
				{
					method: "GET",
					path:   "/v1/projects/test-project/locations/test-location/ragCorpora/1/ragFiles",
				},
			},
		},
		{
			name:     "Success: List files with paging options",
			corpusId: 1,
			options: []ListFilesOption{
				WithListFilesPageSize(10),
				WithListFilesPageToken("next-page-token"),
			},
			setup: func(mt *mockTransport) {
				mt.responses = map[string]mockResponse{
					"GET /v1/projects/test-project/locations/test-location/ragCorpora/1/ragFiles": {
						statusCode: http.StatusOK,
						body: map[string]interface{}{
							"ragFiles": []map[string]interface{}{
								{
									"name":        "projects/test-project/locations/test-location/ragCorpora/1/ragFiles/456",
									"displayName": "Test File 2",
									"description": "Test Description 2",
								},
							},
							"nextPageToken": "next-next-page-token",
						},
					},
				}
			},
			expectedResult: ListFilesResult{
				Files: []File{
					{
						Name:        "projects/test-project/locations/test-location/ragCorpora/1/ragFiles/456",
						DisplayName: "Test File 2",
						Description: "Test Description 2",
					},
				},
				NextPageToken: "next-next-page-token",
			},
			errorExpected: false,
			expectedReqs: []mockRequest{
				{
					method: "GET",
					path:   "/v1/projects/test-project/locations/test-location/ragCorpora/1/ragFiles",
				},
			},
		},
		{
			name:     "Error: API returns error",
			corpusId: 1,
			options:  []ListFilesOption{},
			setup: func(mt *mockTransport) {
				mt.responses = map[string]mockResponse{
					"GET /v1/projects/test-project/locations/test-location/ragCorpora/1/ragFiles": {
						statusCode: http.StatusInternalServerError,
						body:       map[string]interface{}{},
					},
				}
			},
			expectedResult: ListFilesResult{},
			errorExpected:  true,
			expectedReqs: []mockRequest{
				{
					method: "GET",
					path:   "/v1/projects/test-project/locations/test-location/ragCorpora/1/ragFiles",
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Prepare mock transport
			mt := newMockTransport(t, tt.expectedReqs)
			tt.setup(mt)

			// Create test client
			client := NewClient(&http.Client{Transport: mt}, "test-project", "test-location")

			// Execute test
			result, err := client.ListFiles(context.Background(), tt.corpusId, tt.options...)

			// Verify results
			if tt.errorExpected {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedResult, result)
			}

			// Verify requests
			mt.verify(t)
		})
	}
}
