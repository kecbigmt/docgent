package github

import (
	"context"
	"docgent/internal/domain/data"
	"net/http"
	"testing"

	"github.com/google/go-github/v68/github"
	"github.com/stretchr/testify/assert"
)

func TestSourceRepository_Match(t *testing.T) {
	tests := []struct {
		name string
		uri  string
		want bool
	}{
		{
			name: "GitHubのURLの場合はtrueを返す",
			uri:  "https://github.com/owner/repo/pull/123#issuecomment-456789",
			want: true,
		},
		{
			name: "GitHubのURL以外の場合はfalseを返す",
			uri:  "https://app.slack.com/client/T123/C456/thread/1234567890.123",
			want: false,
		},
	}

	client := github.NewClient(nil)
	repo := NewSourceRepository(client)

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			uri := data.NewURIUnsafe(tt.uri)
			got := repo.Match(uri)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestSourceRepository_Find(t *testing.T) {
	tests := []struct {
		name         string
		uri          string
		setup        func(*mockTransport)
		wantErr      bool
		wantContent  string
		expectedReqs []mockRequest
	}{
		{
			name: "正常系：コメントを取得できる",
			uri:  "https://github.com/owner/repo/pull/123#issuecomment-456789",
			setup: func(mt *mockTransport) {
				mt.responses = map[string]mockResponse{
					"GET /repos/owner/repo/issues/123/comments": {
						statusCode: http.StatusOK,
						body: []github.IssueComment{
							{
								ID:   github.Ptr(int64(123456)),
								User: &github.User{Login: github.Ptr("testuser")},
								Body: github.Ptr("テストコメント1"),
							},
							{
								ID:   github.Ptr(int64(456789)),
								User: &github.User{Login: github.Ptr("testuser")},
								Body: github.Ptr("テストコメント2"),
							},
							{
								ID:   github.Ptr(int64(789101)),
								User: &github.User{Login: github.Ptr("testuser")},
								Body: github.Ptr("テストコメント3"),
							},
						},
					},
				}
			},
			wantErr: false,
			wantContent: `<conversation uri="https://github.com/owner/repo/pull/123#issuecomment-456789">
<message user="testuser">
テストコメント1
</message>
<message user="testuser" highlighted="true">
テストコメント2
</message>
<message user="testuser">
テストコメント3
</message>
</conversation>`,
			expectedReqs: []mockRequest{
				{
					method: "GET",
					path:   "/repos/owner/repo/issues/123/comments",
				},
			},
		},
		{
			name: "異常系：不正なURI",
			uri:  "https://github.com/invalid/url",
			setup: func(mt *mockTransport) {
				mt.responses = map[string]mockResponse{}
			},
			wantErr:      true,
			expectedReqs: []mockRequest{},
		},
		{
			name: "異常系：APIエラー",
			uri:  "https://github.com/owner/repo/pull/123#issuecomment-456789",
			setup: func(mt *mockTransport) {
				mt.responses = map[string]mockResponse{
					"GET /repos/owner/repo/issues/123/comments": {
						statusCode: http.StatusInternalServerError,
						body: &github.ErrorResponse{
							Response: &http.Response{StatusCode: http.StatusInternalServerError},
							Message:  "Internal Server Error",
						},
					},
				}
			},
			wantErr: true,
			expectedReqs: []mockRequest{
				{
					method: "GET",
					path:   "/repos/owner/repo/issues/123/comments",
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
			repo := NewSourceRepository(client)

			// テスト対象の実行
			uri := data.NewURIUnsafe(tt.uri)
			source, err := repo.Find(context.Background(), uri)

			// 検証
			if tt.wantErr {
				assert.Error(t, err)
				return
			}

			assert.NoError(t, err)
			assert.NotNil(t, source)
			assert.Equal(t, tt.wantContent, source.Content())

			// リクエストの検証
			mt.verify(t)
		})
	}
}
