package github

import (
	"context"
	"net/http"
	"testing"

	"docgent/internal/domain/data"

	"github.com/google/go-github/v68/github"
	"github.com/stretchr/testify/assert"
)

func TestFileQueryService_GetFilePath(t *testing.T) {
	tests := []struct {
		name    string
		uri     *data.URI
		service *FileQueryService
		want    string
		wantErr bool
	}{
		{
			name:    "success: valid GitHub URI",
			uri:     data.NewURIUnsafe("https://github.com/owner/repo/blob/abcdef1234567890/path/to/file.go"),
			service: NewFileQueryService(nil, "owner", "repo", "main"),
			want:    "path/to/file.go",
			wantErr: false,
		},
		{
			name:    "success: valid GitHub URI with query parameters",
			uri:     data.NewURIUnsafe("https://github.com/owner/repo/blob/abcdef1234567890/path/to/file.go?foo=bar"),
			service: NewFileQueryService(nil, "owner", "repo", "main"),
			want:    "path/to/file.go",
			wantErr: false,
		},
		{
			name:    "error: invalid GitHub URI format",
			uri:     data.NewURIUnsafe("https://github.com/owner/repo/invalid/abcdef1234567890/path/to/file.go"),
			service: NewFileQueryService(nil, "owner", "repo", "main"),
			want:    "",
			wantErr: true,
		},
		{
			name:    "error: non-GitHub URI",
			uri:     data.NewURIUnsafe("https://example.com/path/to/file"),
			service: NewFileQueryService(nil, "owner", "repo", "main"),
			want:    "",
			wantErr: true,
		},
		{
			name:    "error: mismatched owner",
			uri:     data.NewURIUnsafe("https://github.com/different-owner/repo/blob/abcdef1234567890/path/to/file.go"),
			service: NewFileQueryService(nil, "owner", "repo", "main"),
			want:    "",
			wantErr: true,
		},
		{
			name:    "error: mismatched repo",
			uri:     data.NewURIUnsafe("https://github.com/owner/different-repo/blob/abcdef1234567890/path/to/file.go"),
			service: NewFileQueryService(nil, "owner", "repo", "main"),
			want:    "",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.service.GetFilePath(tt.uri)
			if tt.wantErr {
				assert.Error(t, err)
				assert.Empty(t, got)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.want, got)
			}
		})
	}
}

func TestFileQueryService_GetURI(t *testing.T) {
	tests := []struct {
		name         string
		path         string
		setup        func(*mockTransport)
		want         string
		wantErr      bool
		expectedReqs []mockRequest
	}{
		{
			name: "success: get URI for file path",
			path: "path/to/file.go",
			setup: func(mt *mockTransport) {
				mt.responses = map[string]mockResponse{
					"GET /repos/owner/repo/git/ref/heads/main": {
						statusCode: http.StatusOK,
						body: &github.Reference{
							Object: &github.GitObject{
								SHA: github.Ptr("abcdef1234567890"),
							},
						},
					},
				}
			},
			want:    "https://github.com/owner/repo/blob/abcdef1234567890/path/to/file.go",
			wantErr: false,
			expectedReqs: []mockRequest{
				{
					method: "GET",
					path:   "/repos/owner/repo/git/ref/heads/main",
				},
			},
		},
		{
			name: "success: caching - API is called only once for multiple GetURI calls",
			path: "path/to/file.go",
			setup: func(mt *mockTransport) {
				mt.responses = map[string]mockResponse{
					"GET /repos/owner/repo/git/ref/heads/main": {
						statusCode: http.StatusOK,
						body: &github.Reference{
							Object: &github.GitObject{
								SHA: github.Ptr("abcdef1234567890"),
							},
						},
					},
				}
			},
			want:    "https://github.com/owner/repo/blob/abcdef1234567890/path/to/file.go",
			wantErr: false,
			// We expect only one request even though we'll call GetURI twice
			expectedReqs: []mockRequest{
				{
					method: "GET",
					path:   "/repos/owner/repo/git/ref/heads/main",
				},
			},
		},
		{
			name: "error: failed to get ref",
			path: "path/to/file.go",
			setup: func(mt *mockTransport) {
				mt.responses = map[string]mockResponse{
					"GET /repos/owner/repo/git/ref/heads/main": {
						statusCode: http.StatusInternalServerError,
						body: &github.ErrorResponse{
							Response: &http.Response{StatusCode: http.StatusInternalServerError},
							Message:  "Internal Server Error",
						},
					},
				}
			},
			want:    "",
			wantErr: true,
			expectedReqs: []mockRequest{
				{
					method: "GET",
					path:   "/repos/owner/repo/git/ref/heads/main",
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mt := &mockTransport{
				responses:    make(map[string]mockResponse),
				expectedReqs: tt.expectedReqs,
			}
			tt.setup(mt)

			client := github.NewClient(&http.Client{Transport: mt})
			service := NewFileQueryService(client, "owner", "repo", "main")

			uri, err := service.GetURI(context.Background(), tt.path)
			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, uri)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, uri)
				assert.Equal(t, tt.want, uri.String())
			}

			mt.verify(t)

			// For the caching test, call GetURI a second time and verify no additional API calls
			if tt.name == "success: caching - API is called only once for multiple GetURI calls" {
				// Call GetURI again with the same path
				uri2, err := service.GetURI(context.Background(), tt.path)
				assert.NoError(t, err)
				assert.NotNil(t, uri2)
				assert.Equal(t, tt.want, uri2.String())

				// Verify that no additional API calls were made
				mt.verify(t)

				// Call GetURI with a different path, should still use the cached commit SHA
				differentPath := "another/path/file.txt"
				uri3, err := service.GetURI(context.Background(), differentPath)
				assert.NoError(t, err)
				assert.NotNil(t, uri3)
				assert.Equal(t, "https://github.com/owner/repo/blob/abcdef1234567890/"+differentPath, uri3.String())

				// Verify that no additional API calls were made
				mt.verify(t)
			}
		})
	}
}
