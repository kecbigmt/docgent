package lib

import (
	"context"
	"fmt"
	"io"
	"net/http"
)

// DeleteCorpus deletes a RAG corpus by ID.
// References:
// - Example: https://cloud.google.com/vertex-ai/generative-ai/docs/model-reference/rag-api-v1#delete-a-rag-corpus-example-api
// - Parameters: https://cloud.google.com/vertex-ai/generative-ai/docs/model-reference/rag-api-v1#delete-a-rag-corpus-params-api
func (c *Client) DeleteCorpus(ctx context.Context, corpusId int64) error {
	url := fmt.Sprintf("https://%s-aiplatform.googleapis.com/v1/projects/%s/locations/%s/ragCorpora/%d", c.location, c.projectID, c.location, corpusId)
	req, err := http.NewRequestWithContext(ctx, "DELETE", url, nil)
	if err != nil {
		return err
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return err
	}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("failed to delete corpus: %s %s", resp.Status, string(body))
	}

	return nil
}
