package slack

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestConversationService_GetURI(t *testing.T) {
	tests := []struct {
		name string
		ref  *ConversationRef
		want string
	}{
		{
			name: "スレッドの最初のメッセージの場合",
			ref: NewConversationRef(
				"T123456",
				"C789012",
				"1234567890.123456",
				"1234567890.123456",
			),
			want: "https://app.slack.com/client/T123456/C789012/1234567890.123456",
		},
		{
			name: "スレッド内の返信メッセージの場合",
			ref: NewConversationRef(
				"T123456",
				"C789012",
				"1234567890.123456",
				"1234567890.654321",
			),
			want: "https://app.slack.com/client/T123456/C789012/thread/C789012-1234567890.123456/1234567890.654321",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			service := &ConversationService{
				ref: tt.ref,
			}
			got := service.URI()
			assert.Equal(t, tt.want, got.String())
		})
	}
}
