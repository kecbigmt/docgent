package rag

import (
	"context"
	"fmt"
)

func (c *Corpus) DeleteFile(ctx context.Context, fileID int64) error {
	err := c.client.DeleteFile(ctx, c.corpusId, fileID)
	if err != nil {
		return fmt.Errorf("failed to delete file: %w", err)
	}
	return nil
}
