package slack

import (
	"testing"

	"docgent/internal/domain/data"

	"github.com/stretchr/testify/assert"
)

func TestParseConversationRef(t *testing.T) {
	tests := []struct {
		name    string
		ref     string
		want    *ConversationRef
		wantErr bool
	}{
		{
			name: "スレッドの最初のメッセージのURIの場合",
			ref:  "https://app.slack.com/client/T123456/C789012/1234567890.123456",
			want: NewConversationRef(
				"T123456",
				"C789012",
				"1234567890.123456",
				"1234567890.123456",
			),
			wantErr: false,
		},
		{
			name: "スレッド内の返信メッセージのURIの場合",
			ref:  "https://app.slack.com/client/T123456/C789012/thread/C789012-1234567890.123456/1234567890.654321",
			want: NewConversationRef(
				"T123456",
				"C789012",
				"1234567890.123456",
				"1234567890.654321",
			),
			wantErr: false,
		},
		{
			name:    "不正なURIの場合",
			ref:     "https://invalid.url/T123456/C789012",
			want:    nil,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseConversationRef(data.NewURIUnsafe(tt.ref))
			if tt.wantErr {
				assert.Error(t, err)
				return
			}
			assert.NoError(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}
