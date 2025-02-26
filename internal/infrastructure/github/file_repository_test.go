package github

import (
	"context"
	"encoding/base64"
	"net/http"
	"testing"

	"docgent/internal/domain/data"

	"github.com/google/go-github/v68/github"
	"github.com/stretchr/testify/assert"
)

func TestFileRepository_Create(t *testing.T) {
	tests := []struct {
		name         string
		file         *data.File
		setup        func(*mockTransport)
		wantErr      error
		expectedReqs []mockRequest
	}{
		{
			name: "success: create file with frontmatter",
			file: &data.File{
				Path:    "test.md",
				Content: "Hello, world!",
				SourceURIs: []*data.URI{
					data.NewURIUnsafe("https://slack.com/archives/C01234567/p123456789"),
				},
			},
			setup: func(mt *mockTransport) {
				mt.responses = map[string]mockResponse{
					"GET /repos/owner/repo/contents/test.md": {
						statusCode: http.StatusNotFound,
						body: &github.ErrorResponse{
							Response: &http.Response{StatusCode: http.StatusNotFound},
							Message:  "Not Found",
						},
					},
					"PUT /repos/owner/repo/contents/test.md": {
						statusCode: http.StatusCreated,
						body: github.RepositoryContentResponse{
							Content: &github.RepositoryContent{
								Name:    github.Ptr("test.md"),
								Path:    github.Ptr("test.md"),
								Content: github.Ptr(base64.StdEncoding.EncodeToString([]byte("---\nsources:\n  - https://slack.com/archives/C01234567/p123456789\n---\nHello, world!"))),
							},
						},
					},
				}
			},
			wantErr: nil,
			expectedReqs: []mockRequest{
				{
					method: "GET",
					path:   "/repos/owner/repo/contents/test.md",
				},
				{
					method: "PUT",
					path:   "/repos/owner/repo/contents/test.md",
					body: map[string]interface{}{
						"message": "Create file test.md",
						"content": base64.StdEncoding.EncodeToString([]byte("---\nsources:\n  - https://slack.com/archives/C01234567/p123456789\n---\nHello, world!")),
						"branch":  "main",
					},
				},
			},
		},
		{
			name: "error: file already exists",
			file: &data.File{
				Path:    "test.md",
				Content: "Hello, world!",
				SourceURIs: []*data.URI{
					data.NewURIUnsafe("https://slack.com/archives/C01234567/p123456789"),
				},
			},
			setup: func(mt *mockTransport) {
				mt.responses = map[string]mockResponse{
					"GET /repos/owner/repo/contents/test.md": {
						statusCode: http.StatusOK,
						body: &github.RepositoryContent{
							Name:    github.Ptr("test.md"),
							Path:    github.Ptr("test.md"),
							Content: github.Ptr(base64.StdEncoding.EncodeToString([]byte("existing content"))),
						},
					},
				}
			},
			wantErr: data.ErrFileAlreadyExists,
			expectedReqs: []mockRequest{
				{
					method: "GET",
					path:   "/repos/owner/repo/contents/test.md",
				},
			},
		},
		{
			name: "error: failed to access file",
			file: &data.File{
				Path:    "test.md",
				Content: "Hello, world!",
				SourceURIs: []*data.URI{
					data.NewURIUnsafe("https://slack.com/archives/C01234567/p123456789"),
				},
			},
			setup: func(mt *mockTransport) {
				mt.responses = map[string]mockResponse{
					"GET /repos/owner/repo/contents/test.md": {
						statusCode: http.StatusInternalServerError,
						body: &github.ErrorResponse{
							Response: &http.Response{StatusCode: http.StatusInternalServerError},
							Message:  "Internal Server Error",
						},
					},
				}
			},
			wantErr: data.ErrFailedToAccessFile,
			expectedReqs: []mockRequest{
				{
					method: "GET",
					path:   "/repos/owner/repo/contents/test.md",
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
			repo := NewFileRepository(client, "owner", "repo", "main")

			err := repo.Create(context.Background(), tt.file)
			if tt.wantErr != nil {
				assert.ErrorIs(t, err, tt.wantErr)
			} else {
				assert.NoError(t, err)
			}

			mt.verify(t)
		})
	}
}

func TestFileRepository_Update(t *testing.T) {
	tests := []struct {
		name         string
		file         *data.File
		setup        func(*mockTransport)
		wantErr      error
		expectedReqs []mockRequest
	}{
		{
			name: "success: update file with frontmatter",
			file: &data.File{
				Path:    "test.md",
				Content: "Hello, world!",
				SourceURIs: []*data.URI{
					data.NewURIUnsafe("https://slack.com/archives/C01234567/p123456789"),
				},
			},
			setup: func(mt *mockTransport) {
				mt.responses = map[string]mockResponse{
					"GET /repos/owner/repo/contents/test.md": {
						statusCode: http.StatusOK,
						body: &github.RepositoryContent{
							Name:     github.Ptr("test.md"),
							Path:     github.Ptr("test.md"),
							Content:  github.Ptr(base64.StdEncoding.EncodeToString([]byte("---\nsources:\n  - https://slack.com/archives/C01234567/p123456789\n---\nHello, world!"))),
							SHA:      github.Ptr("sha"),
							Encoding: github.Ptr("base64"),
						},
					},
					"PUT /repos/owner/repo/contents/test.md": {
						statusCode: http.StatusOK,
						body:       github.RepositoryContentResponse{},
					},
				}
			},
			wantErr: nil,
			expectedReqs: []mockRequest{
				{
					method: "GET",
					path:   "/repos/owner/repo/contents/test.md",
				},
				{
					method: "PUT",
					path:   "/repos/owner/repo/contents/test.md",
					body: map[string]interface{}{
						"message": "Update file test.md",
						"content": base64.StdEncoding.EncodeToString([]byte("---\nsources:\n  - https://slack.com/archives/C01234567/p123456789\n---\nHello, world!")),
						"branch":  "main",
						"sha":     "sha",
					},
				},
			},
		},
		{
			name: "error: file not found",
			file: &data.File{
				Path:    "test.md",
				Content: "Hello, world!",
				SourceURIs: []*data.URI{
					data.NewURIUnsafe("https://slack.com/archives/C01234567/p123456789"),
				},
			},
			setup: func(mt *mockTransport) {
				mt.responses = map[string]mockResponse{
					"GET /repos/owner/repo/contents/test.md": {
						statusCode: http.StatusNotFound,
						body: &github.ErrorResponse{
							Response: &http.Response{StatusCode: http.StatusNotFound},
							Message:  "Not Found",
						},
					},
				}
			},
			wantErr: data.ErrFileNotFound,
			expectedReqs: []mockRequest{
				{
					method: "GET",
					path:   "/repos/owner/repo/contents/test.md",
				},
			},
		},
		{
			name: "error: failed to access file",
			file: &data.File{
				Path:    "test.md",
				Content: "Hello, world!",
				SourceURIs: []*data.URI{
					data.NewURIUnsafe("https://slack.com/archives/C01234567/p123456789"),
				},
			},
			setup: func(mt *mockTransport) {
				mt.responses = map[string]mockResponse{
					"GET /repos/owner/repo/contents/test.md": {
						statusCode: http.StatusInternalServerError,
						body: &github.ErrorResponse{
							Response: &http.Response{StatusCode: http.StatusInternalServerError},
							Message:  "Internal Server Error",
						},
					},
				}
			},
			wantErr: data.ErrFailedToAccessFile,
			expectedReqs: []mockRequest{
				{
					method: "GET",
					path:   "/repos/owner/repo/contents/test.md",
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
			repo := NewFileRepository(client, "owner", "repo", "main")

			err := repo.Update(context.Background(), tt.file)
			if tt.wantErr != nil {
				assert.ErrorIs(t, err, tt.wantErr)
			} else {
				assert.NoError(t, err)
			}

			mt.verify(t)
		})
	}
}

func TestFileRepository_Get(t *testing.T) {
	tests := []struct {
		name         string
		path         string
		setup        func(*mockTransport)
		want         *data.File
		wantErr      error
		expectedReqs []mockRequest
	}{
		{
			name: "success: get file with frontmatter",
			path: "test.md",
			setup: func(mt *mockTransport) {
				mt.responses = map[string]mockResponse{
					"GET /repos/owner/repo/contents/test.md": {
						statusCode: http.StatusOK,
						body: &github.RepositoryContent{
							Name:     github.Ptr("test.md"),
							Path:     github.Ptr("test.md"),
							Content:  github.Ptr(base64.StdEncoding.EncodeToString([]byte("---\nsources:\n  - https://slack.com/archives/C01234567/p123456789\n---\nHello, world!"))),
							SHA:      github.Ptr("sha"),
							Encoding: github.Ptr("base64"),
						},
					},
				}
			},
			want: &data.File{
				Path:    "test.md",
				Content: "Hello, world!",
				SourceURIs: []*data.URI{
					data.NewURIUnsafe("https://slack.com/archives/C01234567/p123456789"),
				},
			},
			wantErr: nil,
			expectedReqs: []mockRequest{
				{
					method: "GET",
					path:   "/repos/owner/repo/contents/test.md",
				},
			},
		},
		{
			name: "error: file not found",
			path: "test.md",
			setup: func(mt *mockTransport) {
				mt.responses = map[string]mockResponse{
					"GET /repos/owner/repo/contents/test.md": {
						statusCode: http.StatusNotFound,
						body: &github.ErrorResponse{
							Response: &http.Response{StatusCode: http.StatusNotFound},
							Message:  "Not Found",
						},
					},
				}
			},
			want:    nil,
			wantErr: data.ErrFileNotFound,
			expectedReqs: []mockRequest{
				{
					method: "GET",
					path:   "/repos/owner/repo/contents/test.md",
				},
			},
		},
		{
			name: "error: invalid frontmatter",
			path: "test.md",
			setup: func(mt *mockTransport) {
				mt.responses = map[string]mockResponse{
					"GET /repos/owner/repo/contents/test.md": {
						statusCode: http.StatusOK,
						body: &github.RepositoryContent{
							Name:     github.Ptr("test.md"),
							Path:     github.Ptr("test.md"),
							Content:  github.Ptr(base64.StdEncoding.EncodeToString([]byte("---\ninvalid: frontmatter: format\n---\nHello, world!"))),
							SHA:      github.Ptr("sha"),
							Encoding: github.Ptr("base64"),
						},
					},
				}
			},
			want:    nil,
			wantErr: data.ErrInvalidFrontmatter,
			expectedReqs: []mockRequest{
				{
					method: "GET",
					path:   "/repos/owner/repo/contents/test.md",
				},
			},
		},
		{
			name: "error: failed to access file",
			path: "test.md",
			setup: func(mt *mockTransport) {
				mt.responses = map[string]mockResponse{
					"GET /repos/owner/repo/contents/test.md": {
						statusCode: http.StatusInternalServerError,
						body: &github.ErrorResponse{
							Response: &http.Response{StatusCode: http.StatusInternalServerError},
							Message:  "Internal Server Error",
						},
					},
				}
			},
			want:    nil,
			wantErr: data.ErrFailedToAccessFile,
			expectedReqs: []mockRequest{
				{
					method: "GET",
					path:   "/repos/owner/repo/contents/test.md",
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
			repo := NewFileRepository(client, "owner", "repo", "main")

			got, err := repo.Get(context.Background(), tt.path)
			if tt.wantErr != nil {
				assert.ErrorIs(t, err, tt.wantErr)
				assert.Nil(t, got)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.want, got)
			}

			mt.verify(t)
		})
	}
}

func TestFileRepository_Delete(t *testing.T) {
	tests := []struct {
		name         string
		path         string
		setup        func(*mockTransport)
		wantErr      error
		expectedReqs []mockRequest
	}{
		{
			name: "success: delete file",
			path: "test.md",
			setup: func(mt *mockTransport) {
				mt.responses = map[string]mockResponse{
					"GET /repos/owner/repo/contents/test.md": {
						statusCode: http.StatusOK,
						body: &github.RepositoryContent{
							Name:    github.Ptr("test.md"),
							Path:    github.Ptr("test.md"),
							Content: github.Ptr(base64.StdEncoding.EncodeToString([]byte("---\nsources:\n  - https://slack.com/archives/C01234567/p123456789\n---\nHello, world!"))),
							SHA:     github.Ptr("sha"),
						},
					},
					"DELETE /repos/owner/repo/contents/test.md": {
						statusCode: http.StatusOK,
						body: github.RepositoryContentResponse{
							Content: &github.RepositoryContent{
								Name: github.Ptr("test.md"),
								Path: github.Ptr("test.md"),
								SHA:  github.Ptr("sha"),
							},
						},
					},
				}
			},
			wantErr: nil,
			expectedReqs: []mockRequest{
				{
					method: "GET",
					path:   "/repos/owner/repo/contents/test.md",
				},
				{
					method: "DELETE",
					path:   "/repos/owner/repo/contents/test.md",
					body: map[string]interface{}{
						"message": "Delete file test.md",
						"branch":  "main",
						"sha":     "sha",
						"content": nil,
					},
				},
			},
		},
		{
			name: "error: file not found",
			path: "test.md",
			setup: func(mt *mockTransport) {
				mt.responses = map[string]mockResponse{
					"GET /repos/owner/repo/contents/test.md": {
						statusCode: http.StatusNotFound,
						body: &github.ErrorResponse{
							Response: &http.Response{StatusCode: http.StatusNotFound},
							Message:  "Not Found",
						},
					},
				}
			},
			wantErr: data.ErrFileNotFound,
			expectedReqs: []mockRequest{
				{
					method: "GET",
					path:   "/repos/owner/repo/contents/test.md",
				},
			},
		},
		{
			name: "error: failed to access file",
			path: "test.md",
			setup: func(mt *mockTransport) {
				mt.responses = map[string]mockResponse{
					"GET /repos/owner/repo/contents/test.md": {
						statusCode: http.StatusInternalServerError,
						body: &github.ErrorResponse{
							Response: &http.Response{StatusCode: http.StatusInternalServerError},
							Message:  "Internal Server Error",
						},
					},
				}
			},
			wantErr: data.ErrFailedToAccessFile,
			expectedReqs: []mockRequest{
				{
					method: "GET",
					path:   "/repos/owner/repo/contents/test.md",
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
			repo := NewFileRepository(client, "owner", "repo", "main")

			err := repo.Delete(context.Background(), tt.path)
			if tt.wantErr != nil {
				assert.ErrorIs(t, err, tt.wantErr)
			} else {
				assert.NoError(t, err)
			}

			mt.verify(t)
		})
	}
}
