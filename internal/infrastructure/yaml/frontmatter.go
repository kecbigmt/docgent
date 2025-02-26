package yaml

import (
	"fmt"
	"strings"

	"docgent/internal/domain/data"

	"gopkg.in/yaml.v3"
)

type Frontmatter struct {
	Sources []string `yaml:"sources"`
}

// GenerateFrontmatter は知識源のリストからYAMLフロントマターを生成します
func GenerateFrontmatter(sources []*data.URI) (string, error) {
	frontmatter := Frontmatter{
		Sources: make([]string, len(sources)),
	}
	for i, source := range sources {
		frontmatter.Sources[i] = source.Value()
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
func ParseFrontmatter(frontmatter string) ([]*data.URI, error) {
	if frontmatter == "" {
		return []*data.URI{}, nil
	}

	var fm Frontmatter
	if err := yaml.Unmarshal([]byte(frontmatter), &fm); err != nil {
		return nil, fmt.Errorf("failed to unmarshal frontmatter: %w", err)
	}

	sources := make([]*data.URI, len(fm.Sources))
	for i, uri := range fm.Sources {
		source, err := data.NewURI(uri)
		if err != nil {
			return nil, fmt.Errorf("failed to parse uri: %w", err)
		}
		sources[i] = source
	}

	return sources, nil
}

// SplitContentAndFrontmatter はファイル内容からフロントマターと本文を分離します
func SplitContentAndFrontmatter(content string) (frontmatter, body string) {
	parts := strings.SplitN(content, "---\n", 3)
	if len(parts) != 3 {
		return "", content
	}
	return parts[1], parts[2]
}

// CombineContentAndFrontmatter はフロントマターと本文を結合します
func CombineContentAndFrontmatter(frontmatter, body string) string {
	frontmatter = strings.TrimSpace(frontmatter)
	return fmt.Sprintf("---\n%s\n---\n%s", frontmatter, body)
}
