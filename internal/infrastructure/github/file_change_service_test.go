package github

import (
	"context"
	"encoding/base64"
	"net/http"
	"testing"

	"github.com/google/go-github/v68/github"
	"github.com/stretchr/testify/assert"

	"docgent/internal/domain/tooluse"
)

func TestFileChangeService_ModifyFile(t *testing.T) {
	tests := []struct {
		name         string
		setup        func(*mockTransport)
		path         string
		hunks        []tooluse.Hunk
		wantErr      error
		expectedReqs []mockRequest
	}{
		{
			name: "正常系: ファイル更新成功",
			setup: func(mt *mockTransport) {
				mt.responses = map[string]mockResponse{
					"GET /repos/owner/repo/contents/test.txt": {
						statusCode: http.StatusOK,
						body: &github.RepositoryContent{
							Content:  github.Ptr(base64.StdEncoding.EncodeToString([]byte("Hello, World!"))),
							SHA:      github.Ptr("sha"),
							Encoding: github.Ptr("base64"),
						},
					},
					"PUT /repos/owner/repo/contents/test.txt": {
						statusCode: http.StatusOK,
						body:       &github.RepositoryContentResponse{},
					},
				}
			},
			path: "test.txt",
			hunks: []tooluse.Hunk{
				{Search: "World", Replace: "Go"},
			},
			wantErr: nil,
			expectedReqs: []mockRequest{
				{
					method: "GET",
					path:   "/repos/owner/repo/contents/test.txt",
				},
				{
					method: "PUT",
					path:   "/repos/owner/repo/contents/test.txt",
					body: map[string]interface{}{
						"message": "Update file test.txt",
						"content": base64.StdEncoding.EncodeToString([]byte("Hello, Go!")),
						"sha":     "sha",
						"branch":  "main",
					},
				},
			},
		},
		{
			name: "異常系: ファイルが存在しない",
			setup: func(mt *mockTransport) {
				mt.responses = map[string]mockResponse{
					"GET /repos/owner/repo/contents/notfound.txt": {
						statusCode: http.StatusNotFound,
						body:       &github.ErrorResponse{},
					},
				}
			},
			path: "notfound.txt",
			hunks: []tooluse.Hunk{
				{Search: "World", Replace: "Go"},
			},
			wantErr: ErrNotFound,
			expectedReqs: []mockRequest{
				{
					method: "GET",
					path:   "/repos/owner/repo/contents/notfound.txt",
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
			h := NewFileChangeService(client, "owner", "repo", "main")

			err := h.ModifyFile(context.Background(), tt.path, tt.hunks)
			if tt.wantErr != nil {
				assert.ErrorIs(t, err, tt.wantErr)
			} else {
				assert.NoError(t, err)
			}

			mt.verify(t)
		})
	}
}

func TestFileChangeService_RenameFile(t *testing.T) {
	tests := []struct {
		name         string
		setup        func(*mockTransport)
		oldPath      string
		newPath      string
		hunks        []tooluse.Hunk
		wantErr      error
		expectedReqs []mockRequest
	}{
		{
			name: "正常系: ファイルリネーム成功",
			setup: func(mt *mockTransport) {
				mt.responses = map[string]mockResponse{
					"GET /repos/owner/repo/contents/old.txt": {
						statusCode: http.StatusOK,
						body: &github.RepositoryContent{
							Content:  github.Ptr(base64.StdEncoding.EncodeToString([]byte("Hello, World!"))),
							SHA:      github.Ptr("sha"),
							Encoding: github.Ptr("base64"),
						},
					},
					"PUT /repos/owner/repo/contents/new.txt": {
						statusCode: http.StatusOK,
						body:       &github.RepositoryContentResponse{},
					},
					"DELETE /repos/owner/repo/contents/old.txt": {
						statusCode: http.StatusOK,
						body:       &github.RepositoryContentResponse{},
					},
				}
			},
			oldPath: "old.txt",
			newPath: "new.txt",
			hunks: []tooluse.Hunk{
				{Search: "World", Replace: "Go"},
			},
			wantErr: nil,
			expectedReqs: []mockRequest{
				{
					method: "GET",
					path:   "/repos/owner/repo/contents/old.txt",
				},
				{
					method: "PUT",
					path:   "/repos/owner/repo/contents/new.txt",
					body: map[string]interface{}{
						"message": "Create file new.txt",
						"content": base64.StdEncoding.EncodeToString([]byte("Hello, Go!")),
						"branch":  "main",
					},
				},
				{
					method: "DELETE",
					path:   "/repos/owner/repo/contents/old.txt",
					body: map[string]interface{}{
						"message": "Delete file old.txt",
						"sha":     "sha",
						"branch":  "main",
						"content": nil,
					},
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
			h := NewFileChangeService(client, "owner", "repo", "main")

			err := h.RenameFile(context.Background(), tt.oldPath, tt.newPath, tt.hunks)
			if tt.wantErr != nil {
				assert.ErrorIs(t, err, tt.wantErr)
			} else {
				assert.NoError(t, err)
			}

			mt.verify(t)
		})
	}
}

func TestFileChangeService_CreateFile(t *testing.T) {
	tests := []struct {
		name         string
		setup        func(*mockTransport)
		path         string
		content      string
		wantErr      error
		expectedReqs []mockRequest
	}{
		{
			name: "正常系: ファイル作成成功",
			setup: func(mt *mockTransport) {
				mt.responses = map[string]mockResponse{
					"PUT /repos/owner/repo/contents/test.txt": {
						statusCode: http.StatusOK,
						body:       &github.RepositoryContentResponse{},
					},
				}
			},
			path:    "test.txt",
			content: "Hello, World!",
			wantErr: nil,
			expectedReqs: []mockRequest{
				{
					method: "PUT",
					path:   "/repos/owner/repo/contents/test.txt",
					body: map[string]interface{}{
						"message": "Create file test.txt",
						"content": base64.StdEncoding.EncodeToString([]byte("Hello, World!")),
						"branch":  "main",
					},
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
			h := NewFileChangeService(client, "owner", "repo", "main")

			err := h.CreateFile(context.Background(), tt.path, tt.content)
			if tt.wantErr != nil {
				assert.ErrorIs(t, err, tt.wantErr)
			} else {
				assert.NoError(t, err)
			}

			mt.verify(t)
		})
	}
}

func TestFileChangeService_DeleteFile(t *testing.T) {
	tests := []struct {
		name         string
		setup        func(*mockTransport)
		path         string
		wantErr      error
		expectedReqs []mockRequest
	}{
		{
			name: "正常系: ファイル削除成功",
			setup: func(mt *mockTransport) {
				mt.responses = map[string]mockResponse{
					"GET /repos/owner/repo/contents/test.txt": {
						statusCode: http.StatusOK,
						body: &github.RepositoryContent{
							SHA:      github.Ptr("sha"),
							Encoding: github.Ptr("base64"),
						},
					},
					"DELETE /repos/owner/repo/contents/test.txt": {
						statusCode: http.StatusOK,
						body:       &github.RepositoryContentResponse{},
					},
				}
			},
			path:    "test.txt",
			wantErr: nil,
			expectedReqs: []mockRequest{
				{
					method: "GET",
					path:   "/repos/owner/repo/contents/test.txt",
				},
				{
					method: "DELETE",
					path:   "/repos/owner/repo/contents/test.txt",
					body: map[string]interface{}{
						"message": "Delete file test.txt",
						"sha":     "sha",
						"branch":  "main",
						"content": nil,
					},
				},
			},
		},
		{
			name: "異常系: ファイルが存在しない",
			setup: func(mt *mockTransport) {
				mt.responses = map[string]mockResponse{
					"GET /repos/owner/repo/contents/notfound.txt": {
						statusCode: http.StatusNotFound,
						body:       &github.ErrorResponse{},
					},
				}
			},
			path:    "notfound.txt",
			wantErr: ErrNotFound,
			expectedReqs: []mockRequest{
				{
					method: "GET",
					path:   "/repos/owner/repo/contents/notfound.txt",
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
			h := NewFileChangeService(client, "owner", "repo", "main")

			err := h.DeleteFile(context.Background(), tt.path)
			if tt.wantErr != nil {
				assert.ErrorIs(t, err, tt.wantErr)
			} else {
				assert.NoError(t, err)
			}

			mt.verify(t)
		})
	}
}
