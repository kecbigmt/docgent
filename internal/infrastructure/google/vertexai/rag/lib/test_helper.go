package lib

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

type mockTransport struct {
	t            *testing.T
	responses    map[string]mockResponse
	requests     []mockRequest
	expectedReqs []mockRequest
}

type mockResponse struct {
	statusCode int
	body       interface{}
}

type mockRequest struct {
	method                string
	path                  string
	body                  interface{}
	validateMultipartForm func(*testing.T, *http.Request)
}

func newMockTransport(t *testing.T, expectedReqs []mockRequest) *mockTransport {
	return &mockTransport{
		t:            t,
		responses:    make(map[string]mockResponse),
		expectedReqs: expectedReqs,
	}
}

func (m *mockTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	key := req.Method + " " + req.URL.Path

	// Record request
	request := mockRequest{
		method: req.Method,
		path:   req.URL.Path,
	}

	// Special handling for multipart/form-data
	if strings.HasPrefix(req.Header.Get("Content-Type"), "multipart/form-data") {
		for _, expected := range m.expectedReqs {
			if expected.method == req.Method && expected.path == req.URL.Path && expected.validateMultipartForm != nil {
				expected.validateMultipartForm(m.t, req)
			}
		}
	} else {
		// Handle regular JSON request
		if req.Body != nil {
			body, err := io.ReadAll(req.Body)
			if err != nil {
				return nil, err
			}
			req.Body = io.NopCloser(bytes.NewReader(body))

			if len(body) > 0 {
				var v interface{}
				if err := json.Unmarshal(body, &v); err != nil {
					return nil, err
				}
				request.body = v
			}
		}
	}

	m.requests = append(m.requests, request)

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
	assert.Equal(t, len(m.expectedReqs), len(m.requests), "Number of requests does not match")
	for i, expected := range m.expectedReqs {
		if i >= len(m.requests) {
			t.Errorf("Expected request %d was not executed: %+v", i, expected)
			continue
		}
		actual := m.requests[i]
		assert.Equal(t, expected.method, actual.method, fmt.Sprintf("Method does not match for request %d", i))
		assert.Equal(t, expected.path, actual.path, fmt.Sprintf("Path does not match for request %d", i))
		if expected.body != nil {
			assert.Equal(t, expected.body, actual.body, fmt.Sprintf("Body does not match for request %d", i))
		}
	}
}
