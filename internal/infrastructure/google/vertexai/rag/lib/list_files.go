package lib

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
)

type ListFilesResponse struct {
	RagFiles      []File `json:"ragFiles"`
	NextPageToken string `json:"nextPageToken,omitempty"`
}

type ListFilesResult struct {
	Files         []File `json:"files"`
	NextPageToken string `json:"nextPageToken,omitempty"`
}

func (c *Client) ListFiles(ctx context.Context, corpusId int64, options ...ListFilesOption) (ListFilesResult, error) {
	listFilesOptions := &ListFilesOptions{}
	for _, option := range options {
		option(listFilesOptions)
	}

	params := url.Values{}
	if listFilesOptions.PageSize != 0 {
		params.Set("pageSize", strconv.Itoa(listFilesOptions.PageSize))
	}
	if listFilesOptions.PageToken != "" {
		params.Set("pageToken", listFilesOptions.PageToken)
	}

	url := fmt.Sprintf("https://%s-aiplatform.googleapis.com/v1/projects/%s/locations/%s/ragCorpora/%d/ragFiles", c.location, c.projectID, c.location, corpusId)
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return ListFilesResult{}, err
	}

	req.URL.RawQuery = params.Encode()

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return ListFilesResult{}, err
	}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return ListFilesResult{}, fmt.Errorf("failed to list files: %d", resp.StatusCode)
	}

	var response ListFilesResponse
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return ListFilesResult{}, err
	}

	return ListFilesResult{
		Files:         response.RagFiles,
		NextPageToken: response.NextPageToken,
	}, nil
}

type ListFilesOption func(*ListFilesOptions)

type ListFilesOptions struct {
	PageSize  int
	PageToken string
}

func WithListFilesPageSize(pageSize int) ListFilesOption {
	return func(o *ListFilesOptions) {
		o.PageSize = pageSize
	}
}

func WithListFilesPageToken(pageToken string) ListFilesOption {
	return func(o *ListFilesOptions) {
		o.PageToken = pageToken
	}
}
