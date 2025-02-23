package github

import (
	"bytes"
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
