package tooluse

import (
	"fmt"
	"strings"
)

// generateFrontmatter は知識源URIのリストからYAMLフロントマターを生成します
func generateFrontmatter(uris []string) string {
	var b strings.Builder
	b.WriteString("knowledge_sources:\n")
	for _, uri := range uris {
		b.WriteString(fmt.Sprintf("  - %q\n", uri))
	}
	return b.String()
}

// splitFrontmatterAndContent はファイルの内容からYAMLフロントマターと本文を分離します
func splitFrontmatterAndContent(content string) (string, string, error) {
	parts := strings.SplitN(content, "---\n", 3)
	if len(parts) != 3 {
		return "", content, nil
	}
	return parts[1], parts[2], nil
}

// extractContent はファイルの内容からフロントマターを除いた本文を抽出します
func extractContent(content string) string {
	_, content, _ = splitFrontmatterAndContent(content)
	return content
}
