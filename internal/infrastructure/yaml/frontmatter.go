package yaml

import (
	"fmt"
	"strings"

	"docgent/internal/domain/data"

	"gopkg.in/yaml.v3"
)

type Frontmatter struct {
	KnowledgeSources []string `yaml:"knowledge_sources"`
}

// GenerateFrontmatter は知識源のリストからYAMLフロントマターを生成します
func GenerateFrontmatter(sources []data.KnowledgeSource) (string, error) {
	frontmatter := Frontmatter{
		KnowledgeSources: make([]string, len(sources)),
	}
	for i, source := range sources {
		frontmatter.KnowledgeSources[i] = source.URI
	}

	var buf strings.Builder
	encoder := yaml.NewEncoder(&buf)
	encoder.SetIndent(2) // インデントを2スペースに設定

	if err := encoder.Encode(frontmatter); err != nil {
		return "", fmt.Errorf("failed to marshal frontmatter: %w", err)
	}

	return buf.String(), nil
}

// ParseFrontmatter はYAMLフロントマターから知識源情報を抽出します
func ParseFrontmatter(frontmatter string) ([]data.KnowledgeSource, error) {
	var fm Frontmatter
	if err := yaml.Unmarshal([]byte(frontmatter), &fm); err != nil {
		return nil, fmt.Errorf("failed to unmarshal frontmatter: %w", err)
	}

	sources := make([]data.KnowledgeSource, len(fm.KnowledgeSources))
	for i, uri := range fm.KnowledgeSources {
		sources[i] = data.KnowledgeSource{URI: uri}
	}

	return sources, nil
}

// SplitContentAndFrontmatter はファイル内容からフロントマターと本文を分離します
func SplitContentAndFrontmatter(content string) (frontmatter, body string, err error) {
	parts := strings.SplitN(content, "---\n", 3)
	if len(parts) != 3 {
		return "", content, nil
	}
	return parts[1], parts[2], nil
}

// CombineContentAndFrontmatter はフロントマターと本文を結合します
func CombineContentAndFrontmatter(frontmatter, body string) string {
	frontmatter = strings.TrimSpace(frontmatter)
	return fmt.Sprintf("---\n%s\n---\n%s", frontmatter, body)
}
