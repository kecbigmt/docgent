package lib

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

// CreateCorpus creates a new RAG corpus.
// References:
// - Example: https://cloud.google.com/vertex-ai/generative-ai/docs/model-reference/rag-api-v1#create-a-rag-corpus-example-api
// - Parameters: https://cloud.google.com/vertex-ai/generative-ai/docs/model-reference/rag-api-v1#corpus-management-params-api
func (c *Client) CreateCorpus(ctx context.Context, displayName string, options ...CreateCorpusOption) error {
	createCorpusOptions := &CreateCorpusOptions{}
	for _, option := range options {
		option(createCorpusOptions)
	}

	reqBody := map[string]interface{}{
		"display_name": displayName,
	}

	if createCorpusOptions.Description != "" {
		reqBody["description"] = createCorpusOptions.Description
	}

	reqBodyBytes, err := json.Marshal(reqBody)
	if err != nil {
		return err
	}

	url := fmt.Sprintf("https://%s-aiplatform.googleapis.com/v1/projects/%s/locations/%s/ragCorpora", c.location, c.projectID, c.location)
	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(reqBodyBytes))
	if err != nil {
		return err
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return err
	}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("failed to create corpus: %s %s", resp.Status, string(body))
	}

	return nil
}

type CreateCorpusOption func(*CreateCorpusOptions)

type CreateCorpusOptions struct {
	Description string
}

func WithCreateCorpusDescription(description string) CreateCorpusOption {
	return func(o *CreateCorpusOptions) {
		o.Description = description
	}
}
