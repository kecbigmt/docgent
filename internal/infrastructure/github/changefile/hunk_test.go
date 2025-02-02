package changefile

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"docgent-backend/internal/domain/tooluse"
)

func TestApplyHunks(t *testing.T) {
	tests := []struct {
		name    string
		content string
		hunks   []tooluse.Hunk
		want    string
		wantErr error
	}{
		{
			name:    "正常系: 単一の置換",
			content: "Hello, World!",
			hunks: []tooluse.Hunk{
				{Search: "World", Replace: "Go"},
			},
			want:    "Hello, Go!",
			wantErr: nil,
		},
		{
			name:    "正常系: 空文字列の検索（先頭に追加）",
			content: "Hello, World!",
			hunks: []tooluse.Hunk{
				{Search: "", Replace: "Hi! "},
			},
			want:    "Hi! Hello, World!",
			wantErr: nil,
		},
		{
			name:    "正常系: 複数のHunk",
			content: "Hello, World!",
			hunks: []tooluse.Hunk{
				{Search: "Hello", Replace: "Hi"},
				{Search: "World", Replace: "Go"},
			},
			want:    "Hi, Go!",
			wantErr: nil,
		},
		{
			name:    "異常系: 検索文字列が見つからない",
			content: "Hello, World!",
			hunks: []tooluse.Hunk{
				{Search: "Golang", Replace: "Go"},
			},
			want:    "",
			wantErr: ErrSearchStringNotFound,
		},
		{
			name:    "異常系: 検索文字列が複数回出現",
			content: "Hello, Hello, World!",
			hunks: []tooluse.Hunk{
				{Search: "Hello", Replace: "Hi"},
			},
			want:    "",
			wantErr: ErrMultipleMatches,
		},
		{
			name:    "正常系: コンテキスト付きの置換",
			content: "func main() {\n\tfmt.Println(\"Hello\")\n}",
			hunks: []tooluse.Hunk{
				{Search: "fmt.Println(\"Hello\")", Replace: "fmt.Println(\"Hello, Go!\")"},
			},
			want:    "func main() {\n\tfmt.Println(\"Hello, Go!\")\n}",
			wantErr: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := applyHunks(tt.content, tt.hunks)
			if tt.wantErr != nil {
				assert.ErrorIs(t, err, tt.wantErr)
				return
			}
			assert.NoError(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}
