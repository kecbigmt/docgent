package lib

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
)

type ListRagCorporaResponse struct {
	RagCorpora    []Corpus `json:"ragCorpora"`
	NextPageToken string   `json:"nextPageToken,omitempty"`
}

type ListCorporaResult struct {
	Corpora       []Corpus `json:"corpora"`
	NextPageToken string   `json:"nextPageToken,omitempty"`
}

// ListCorpora lists the RAG corpora in the project.
// References:
// - Example: https://cloud.google.com/vertex-ai/generative-ai/docs/model-reference/rag-api-v1#list-rag-files-example-api
// - Parameters: https://cloud.google.com/vertex-ai/generative-ai/docs/model-reference/rag-api-v1#list-rag-corpora-params-api
func (c *Client) ListCorpora(ctx context.Context, options ...ListCorpusOption) (ListCorporaResult, error) {
	listCorpusOptions := &ListCorpusOptions{}
	for _, option := range options {
		option(listCorpusOptions)
	}

	params := url.Values{}
	if listCorpusOptions.PageSize != 0 {
		params.Add("pageSize", strconv.Itoa(listCorpusOptions.PageSize))
	}

	if listCorpusOptions.PageToken != "" {
		params.Add("pageToken", listCorpusOptions.PageToken)
	}

	url := fmt.Sprintf("https://%s-aiplatform.googleapis.com/v1/projects/%s/locations/%s/ragCorpora", c.location, c.projectID, c.location)
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return ListCorporaResult{}, err
	}

	req.URL.RawQuery = params.Encode()

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return ListCorporaResult{}, err
	}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return ListCorporaResult{}, &HTTPError{
			StatusCode: resp.StatusCode,
			Status:     resp.Status,
			RawBody:    string(body),
		}
	}

	var response ListRagCorporaResponse
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return ListCorporaResult{}, err
	}

	return ListCorporaResult{
		Corpora:       response.RagCorpora,
		NextPageToken: response.NextPageToken,
	}, nil
}

type ListCorpusOption func(*ListCorpusOptions)

type ListCorpusOptions struct {
	PageSize  int
	PageToken string
}

func WithListCorpusPageSize(pageSize int) ListCorpusOption {
	return func(o *ListCorpusOptions) {
		o.PageSize = pageSize
	}
}

func WithListCorpusPageToken(pageToken string) ListCorpusOption {
	return func(o *ListCorpusOptions) {
		o.PageToken = pageToken
	}
}
