package cli

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"strconv"

	"docgent/internal/infrastructure/google/vertexai/rag/lib"
)

func HandleCorpusRetrieve(ctx context.Context, cli *CLI, client *lib.Client) error {
	corpusID, err := strconv.ParseInt(cli.Corpus.Retrieve.CorpusID, 10, 64)
	if err != nil {
		return fmt.Errorf("invalid corpus ID format: %v", err)
	}

	results, err := client.RetrieveContexts(
		ctx,
		corpusID,
		cli.Corpus.Retrieve.Query,
		cli.Corpus.Retrieve.TopK,
		cli.Corpus.Retrieve.VectorDistanceThreshold,
	)
	if err != nil {
		var httpErr *lib.HTTPError
		if errors.As(err, &httpErr) {
			return fmt.Errorf("検索に失敗しました: %s %s", httpErr.Status, httpErr.RawBody)
		}
		return fmt.Errorf("検索に失敗しました: %v", err)
	}

	fmt.Printf("検索結果: %d件見つかりました\n\n", len(results.Contexts.Contexts))

	// 構造化されたJSON出力
	encoder := json.NewEncoder(os.Stdout)
	encoder.SetIndent("", "  ")
	if err := encoder.Encode(results); err != nil {
		return fmt.Errorf("検索結果のエンコードに失敗しました: %v", err)
	}

	return nil
}
