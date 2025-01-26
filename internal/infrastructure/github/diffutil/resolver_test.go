package diffutil

import (
	"bytes"
	"docgent-backend/internal/domain"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"testing"

	"github.com/google/go-github/v68/github"
)

type mockTransport struct {
	// リクエストに対するレスポンスをマップで保持
	responses map[string]mockResponse
}

type mockResponse struct {
	statusCode int
	body       interface{}
}

func (m *mockTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	key := req.Method + " " + req.URL.Path
	resp, ok := m.responses[key]
	if !ok {
		return &http.Response{
			StatusCode: http.StatusNotFound,
			Body:       io.NopCloser(bytes.NewBufferString(`{"message": "not found"}`)),
			Header:     make(http.Header),
		}, nil
	}

	// リクエストボディの検証
	if req.Body != nil {
		var reqBody map[string]interface{}
		if err := json.NewDecoder(req.Body).Decode(&reqBody); err != nil {
			return nil, fmt.Errorf("failed to decode request body: %w", err)
		}
		req.Body.Close()

		// ファイル作成・更新の場合、必要なフィールドが含まれているか確認
		if req.Method == "PUT" {
			if _, ok := reqBody["message"]; !ok {
				return nil, fmt.Errorf("missing required field 'message'")
			}
			if _, ok := reqBody["content"]; !ok {
				return nil, fmt.Errorf("missing required field 'content'")
			}
			if _, ok := reqBody["branch"]; !ok {
				return nil, fmt.Errorf("missing required field 'branch'")
			}

			// 新規ファイル作成以外の場合はSHAが必須
			if strings.HasSuffix(req.URL.Path, "/file.txt") || strings.HasSuffix(req.URL.Path, "/old.txt") {
				if _, ok := reqBody["sha"]; !ok {
					return nil, fmt.Errorf("missing required field 'sha'")
				}
			}
		}
	}

	body, err := json.Marshal(resp.body)
	if err != nil {
		return nil, err
	}

	return &http.Response{
		StatusCode: resp.statusCode,
		Body:       io.NopCloser(bytes.NewBuffer(body)),
		Header:     make(http.Header),
	}, nil
}

func encodeContent(s string) string {
	return base64.StdEncoding.EncodeToString([]byte(s))
}

func TestResolver_Execute_CreateNewFile(t *testing.T) {
	mock := &mockTransport{
		responses: map[string]mockResponse{
			"PUT /repos/owner/repo/contents/newfile.txt": {
				statusCode: http.StatusCreated,
				body: github.RepositoryContentResponse{
					Content: &github.RepositoryContent{
						Name:    github.Ptr("newfile.txt"),
						Path:    github.Ptr("newfile.txt"),
						Content: github.Ptr(encodeContent("This is a new file.\n")),
					},
				},
			},
		},
	}

	client := github.NewClient(&http.Client{Transport: mock})
	resolver := &Resolver{
		client:     client,
		owner:      "owner",
		repo:       "repo",
		branchName: "main",
	}

	diff := domain.Diff{
		OldName:   "",
		NewName:   "newfile.txt",
		Body:      "@@ -0,0 +1 @@\n+This is a new file.\n",
		IsNewFile: true,
	}

	err := resolver.Execute(diff)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestResolver_Execute_UpdateFileWithoutRename(t *testing.T) {
	mock := &mockTransport{
		responses: map[string]mockResponse{
			"GET /repos/owner/repo/contents/file.txt": {
				statusCode: http.StatusOK,
				body: &github.RepositoryContent{
					Name:    github.Ptr("file.txt"),
					Path:    github.Ptr("file.txt"),
					Content: github.Ptr(encodeContent("Hello\nWorld\n")),
					SHA:     github.Ptr("abc123"),
				},
			},
			"PUT /repos/owner/repo/contents/file.txt": {
				statusCode: http.StatusOK,
				body: github.RepositoryContentResponse{
					Content: &github.RepositoryContent{
						Name:    github.Ptr("file.txt"),
						Path:    github.Ptr("file.txt"),
						Content: github.Ptr(encodeContent("Hi\nWorld\n")),
						SHA:     github.Ptr("def456"),
					},
				},
			},
		},
	}

	client := github.NewClient(&http.Client{Transport: mock})
	resolver := &Resolver{
		client:     client,
		owner:      "owner",
		repo:       "repo",
		branchName: "main",
	}

	diff := domain.Diff{
		OldName:   "file.txt",
		NewName:   "file.txt",
		Body:      "@@ -1,2 +1,2 @@\n-Hello\n+Hi\n World\n",
		IsNewFile: false,
	}

	err := resolver.Execute(diff)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestResolver_Execute_UpdateFileWithRename(t *testing.T) {
	mock := &mockTransport{
		responses: map[string]mockResponse{
			"GET /repos/owner/repo/contents/new.txt": {
				statusCode: http.StatusOK,
				body: &github.RepositoryContent{
					Name:    github.Ptr("new.txt"),
					Path:    github.Ptr("new.txt"),
					Content: github.Ptr("SGVsbG8KV29ybGQK"), // base64 encoded "Hello\nWorld\n"
					SHA:     github.Ptr("abc123"),
				},
			},
			"DELETE /repos/owner/repo/contents/old.txt": {
				statusCode: http.StatusOK,
				body: github.RepositoryContentResponse{
					Content: &github.RepositoryContent{
						Name: github.Ptr("old.txt"),
						Path: github.Ptr("old.txt"),
						SHA:  github.Ptr("abc123"),
					},
				},
			},
			"PUT /repos/owner/repo/contents/new.txt": {
				statusCode: http.StatusOK,
				body: github.RepositoryContentResponse{
					Content: &github.RepositoryContent{
						Name:    github.Ptr("new.txt"),
						Path:    github.Ptr("new.txt"),
						Content: github.Ptr(encodeContent("Hi\nWorld\n")),
						SHA:     github.Ptr("def456"),
					},
				},
			},
		},
	}

	client := github.NewClient(&http.Client{Transport: mock})
	resolver := &Resolver{
		client:     client,
		owner:      "owner",
		repo:       "repo",
		branchName: "main",
	}

	diff := domain.Diff{
		OldName:   "old.txt",
		NewName:   "new.txt",
		Body:      "@@ -1,2 +1,2 @@\n-Hello\n+Hi\n World\n",
		IsNewFile: false,
	}

	err := resolver.Execute(diff)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestResolver_Execute_FileNotFound(t *testing.T) {
	mock := &mockTransport{
		responses: map[string]mockResponse{
			"GET /repos/owner/repo/contents/nonexistent.txt": {
				statusCode: http.StatusNotFound,
				body: &github.ErrorResponse{
					Response: &http.Response{StatusCode: http.StatusNotFound},
					Message:  "Not Found",
				},
			},
		},
	}

	client := github.NewClient(&http.Client{Transport: mock})
	resolver := &Resolver{
		client:     client,
		owner:      "owner",
		repo:       "repo",
		branchName: "main",
	}

	diff := domain.Diff{
		OldName:   "nonexistent.txt",
		NewName:   "nonexistent.txt",
		Body:      "@@ -1,1 +1,1 @@\n-old\n+new\n",
		IsNewFile: false,
	}

	err := resolver.Execute(diff)
	if err == nil {
		t.Error("expected error, got nil")
	}
	if !strings.Contains(err.Error(), "failed to get file content") {
		t.Errorf("expected error to contain 'failed to get file content', got %v", err)
	}
}

func TestResolver_Execute_PermissionDenied(t *testing.T) {
	mock := &mockTransport{
		responses: map[string]mockResponse{
			"PUT /repos/owner/repo/contents/file.txt": {
				statusCode: http.StatusForbidden,
				body: &github.ErrorResponse{
					Response: &http.Response{StatusCode: http.StatusForbidden},
					Message:  "Permission denied",
				},
			},
		},
	}

	client := github.NewClient(&http.Client{Transport: mock})
	resolver := &Resolver{
		client:     client,
		owner:      "owner",
		repo:       "repo",
		branchName: "main",
	}

	diff := domain.Diff{
		OldName:   "",
		NewName:   "file.txt",
		Body:      "@@ -0,0 +1 @@\n+new file\n",
		IsNewFile: true,
	}

	err := resolver.Execute(diff)
	if err == nil {
		t.Error("expected error, got nil")
	}
	if !strings.Contains(err.Error(), "failed to create file") {
		t.Errorf("expected error to contain 'failed to create file', got %v", err)
	}
}

func TestResolver_Execute_InvalidDiff(t *testing.T) {
	client := github.NewClient(&http.Client{Transport: &mockTransport{}})
	resolver := &Resolver{
		client:     client,
		owner:      "owner",
		repo:       "repo",
		branchName: "main",
	}

	diff := domain.Diff{
		OldName:   "file.txt",
		NewName:   "file.txt",
		Body:      "invalid diff format",
		IsNewFile: false,
	}

	err := resolver.Execute(diff)
	if err == nil {
		t.Error("expected error, got nil")
	}
	if !strings.Contains(err.Error(), "failed to parse diff") {
		t.Errorf("expected error to contain 'failed to parse diff', got %v", err)
	}
}
