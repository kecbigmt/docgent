package lib

import (
	"context"
	"fmt"
	"io"
	"net/http"
)

// DeleteFile deletes a file from a RAG corpus. Reference: https://cloud.google.com/vertex-ai/generative-ai/docs/model-reference/rag-api-v1#delete-a-rag-file-example-api
func (c *Client) DeleteFile(ctx context.Context, corpusId int64, ragFileId int64) error {
	url := fmt.Sprintf("https://%s-aiplatform.googleapis.com/v1/projects/%s/locations/%s/ragCorpora/%d/ragFiles/%d", c.location, c.projectID, c.location, corpusId, ragFileId)
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
		return fmt.Errorf("failed to delete file: %s %s", resp.Status, string(body))
	}

	return nil
}
