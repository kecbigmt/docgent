package yaml

import (
	"testing"

	"docgent/internal/domain/data"

	"github.com/stretchr/testify/assert"
)

func TestGenerateFrontmatter(t *testing.T) {
	tests := []struct {
		name          string
		sources       []data.KnowledgeSource
		expected      string
		expectedError bool
	}{
		{
			name: "正常系：単一の知識源",
			sources: []data.KnowledgeSource{
				{URI: "https://slack.com/archives/C01234567/p123456789"},
			},
			expected: "knowledge_sources:\n  - https://slack.com/archives/C01234567/p123456789\n",
		},
		{
			name: "正常系：複数の知識源",
			sources: []data.KnowledgeSource{
				{URI: "https://slack.com/archives/C01234567/p123456789"},
				{URI: "https://github.com/user/repo/pull/1"},
			},
			expected: "knowledge_sources:\n  - https://slack.com/archives/C01234567/p123456789\n  - https://github.com/user/repo/pull/1\n",
		},
		{
			name:     "正常系：空の知識源リスト",
			sources:  []data.KnowledgeSource{},
			expected: "knowledge_sources: []\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := GenerateFrontmatter(tt.sources)
			if tt.expectedError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expected, result)
			}
		})
	}
}

func TestParseFrontmatter(t *testing.T) {
	tests := []struct {
		name          string
		frontmatter   string
		expected      []data.KnowledgeSource
		expectedError bool
	}{
		{
			name:        "正常系：単一の知識源",
			frontmatter: "knowledge_sources:\n  - https://slack.com/archives/C01234567/p123456789\n",
			expected: []data.KnowledgeSource{
				{URI: "https://slack.com/archives/C01234567/p123456789"},
			},
		},
		{
			name:        "正常系：複数の知識源",
			frontmatter: "knowledge_sources:\n  - https://slack.com/archives/C01234567/p123456789\n  - https://github.com/user/repo/pull/1\n",
			expected: []data.KnowledgeSource{
				{URI: "https://slack.com/archives/C01234567/p123456789"},
				{URI: "https://github.com/user/repo/pull/1"},
			},
		},
		{
			name:        "正常系：空の知識源リスト",
			frontmatter: "knowledge_sources: []\n",
			expected:    []data.KnowledgeSource{},
		},
		{
			name:          "正常系：空のフロントマター",
			frontmatter:   "",
			expected:      []data.KnowledgeSource{},
			expectedError: false,
		},
		{
			name:          "エラー系：不正なYAML形式",
			frontmatter:   "invalid: - yaml: format",
			expectedError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := ParseFrontmatter(tt.frontmatter)
			if tt.expectedError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expected, result)
			}
		})
	}
}

func TestSplitContentAndFrontmatter(t *testing.T) {
	tests := []struct {
		name                string
		content             string
		expectedFrontmatter string
		expectedBody        string
	}{
		{
			name:                "正常系：フロントマターあり",
			content:             "---\nknowledge_sources:\n  - https://slack.com/archives/C01234567/p123456789\n---\n# Hello\nWorld",
			expectedFrontmatter: "knowledge_sources:\n  - https://slack.com/archives/C01234567/p123456789\n",
			expectedBody:        "# Hello\nWorld",
		},
		{
			name:                "正常系：フロントマターなし",
			content:             "# Hello\nWorld",
			expectedFrontmatter: "",
			expectedBody:        "# Hello\nWorld",
		},
		{
			name:                "正常系：空のフロントマター",
			content:             "---\n---\n# Hello\nWorld",
			expectedFrontmatter: "",
			expectedBody:        "# Hello\nWorld",
		},
		{
			name:                "正常系：本文なし",
			content:             "---\nknowledge_sources:\n  - https://slack.com/archives/C01234567/p123456789\n---\n",
			expectedFrontmatter: "knowledge_sources:\n  - https://slack.com/archives/C01234567/p123456789\n",
			expectedBody:        "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			frontmatter, body := SplitContentAndFrontmatter(tt.content)
			assert.Equal(t, tt.expectedFrontmatter, frontmatter)
			assert.Equal(t, tt.expectedBody, body)
		})
	}
}

func TestCombineContentAndFrontmatter(t *testing.T) {
	tests := []struct {
		name        string
		frontmatter string
		body        string
		expected    string
	}{
		{
			name:        "正常系：フロントマターと本文あり",
			frontmatter: "knowledge_sources:\n  - https://slack.com/archives/C01234567/p123456789\n",
			body:        "# Hello\nWorld",
			expected:    "---\nknowledge_sources:\n  - https://slack.com/archives/C01234567/p123456789\n---\n# Hello\nWorld",
		},
		{
			name:        "正常系：空のフロントマター",
			frontmatter: "",
			body:        "# Hello\nWorld",
			expected:    "---\n\n---\n# Hello\nWorld",
		},
		{
			name:        "正常系：空の本文",
			frontmatter: "knowledge_sources:\n  - https://slack.com/archives/C01234567/p123456789\n",
			body:        "",
			expected:    "---\nknowledge_sources:\n  - https://slack.com/archives/C01234567/p123456789\n---\n",
		},
		{
			name:        "正常系：両方空",
			frontmatter: "",
			body:        "",
			expected:    "---\n\n---\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := CombineContentAndFrontmatter(tt.frontmatter, tt.body)
			assert.Equal(t, tt.expected, result)
		})
	}
}
