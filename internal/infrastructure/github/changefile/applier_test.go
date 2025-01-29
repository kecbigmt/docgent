package changefile

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"testing"

	"github.com/google/go-github/v68/github"
	"github.com/stretchr/testify/assert"

	"docgent-backend/internal/domain/command"
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

func TestApplier_HandleModify(t *testing.T) {
	tests := []struct {
		name         string
		setup        func(*mockTransport)
		cmd          command.ModifyFile
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
			cmd: command.ModifyFile{
				Path: "test.txt",
				Hunks: []command.Hunk{
					{Search: "World", Replace: "Go"},
				},
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
			cmd: command.ModifyFile{
				Path: "notfound.txt",
				Hunks: []command.Hunk{
					{Search: "World", Replace: "Go"},
				},
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
			h := NewApplier(client, "owner", "repo", "main")

			err := h.handleModify(context.Background(), tt.cmd)
			if tt.wantErr != nil {
				assert.ErrorIs(t, err, tt.wantErr)
			} else {
				assert.NoError(t, err)
			}

			mt.verify(t)
		})
	}
}

func TestApplier_HandleRename(t *testing.T) {
	tests := []struct {
		name         string
		setup        func(*mockTransport)
		cmd          command.RenameFile
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
			cmd: command.RenameFile{
				OldPath: "old.txt",
				NewPath: "new.txt",
				Hunks: []command.Hunk{
					{Search: "World", Replace: "Go"},
				},
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
		{
			name: "異常系: 元ファイルが存在しない",
			setup: func(mt *mockTransport) {
				mt.responses = map[string]mockResponse{
					"GET /repos/owner/repo/contents/notfound.txt": {
						statusCode: http.StatusNotFound,
						body:       &github.ErrorResponse{},
					},
				}
			},
			cmd: command.RenameFile{
				OldPath: "notfound.txt",
				NewPath: "new.txt",
				Hunks:   []command.Hunk{},
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
			h := NewApplier(client, "owner", "repo", "main")

			err := h.handleRename(context.Background(), tt.cmd)
			if tt.wantErr != nil {
				assert.ErrorIs(t, err, tt.wantErr)
			} else {
				assert.NoError(t, err)
			}

			mt.verify(t)
		})
	}
}
