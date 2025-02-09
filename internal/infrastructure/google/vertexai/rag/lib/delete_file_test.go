package lib

import (
	"context"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestClient_DeleteFile(t *testing.T) {
	tests := []struct {
		name          string
		corpusId      int64
		ragFileId     int64
		setup         func(*mockTransport)
		errorExpected bool
		expectedReqs  []mockRequest
	}{
		{
			name:      "Success: Delete file",
			corpusId:  1,
			ragFileId: 123,
			setup: func(mt *mockTransport) {
				mt.responses = map[string]mockResponse{
					"DELETE /v1/projects/test-project/locations/test-location/ragCorpora/1/ragFiles/123": {
						statusCode: http.StatusOK,
						body:       map[string]interface{}{},
					},
				}
			},
			errorExpected: false,
			expectedReqs: []mockRequest{
				{
					method: "DELETE",
					path:   "/v1/projects/test-project/locations/test-location/ragCorpora/1/ragFiles/123",
				},
			},
		},
		{
			name:      "Error: API returns error",
			corpusId:  1,
			ragFileId: 123,
			setup: func(mt *mockTransport) {
				mt.responses = map[string]mockResponse{
					"DELETE /v1/projects/test-project/locations/test-location/ragCorpora/1/ragFiles/123": {
						statusCode: http.StatusInternalServerError,
						body:       map[string]interface{}{},
					},
				}
			},
			errorExpected: true,
			expectedReqs: []mockRequest{
				{
					method: "DELETE",
					path:   "/v1/projects/test-project/locations/test-location/ragCorpora/1/ragFiles/123",
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
			err := client.DeleteFile(
				context.Background(),
				tt.corpusId,
				tt.ragFileId,
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
