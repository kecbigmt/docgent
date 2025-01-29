package changefile

import (
	"fmt"
	"strings"

	"docgent-backend/internal/domain/command"
)

func applyHunks(content string, hunks []command.Hunk) (string, error) {
	result := content
	for _, hunk := range hunks {
		if hunk.Search == "" {
			// 空文字列の場合は出現回数チェックをスキップ
			result = hunk.Replace + result
			continue
		}

		count := strings.Count(result, hunk.Search)
		if count == 0 {
			return "", fmt.Errorf("%w: %q", ErrSearchStringNotFound, hunk.Search)
		}
		if count > 1 {
			return "", fmt.Errorf("%w: multiple occurrences (%d) of search string: %q", ErrMultipleMatches, count, hunk.Search)
		}

		result = strings.ReplaceAll(result, hunk.Search, hunk.Replace)
	}
	return result, nil
}
