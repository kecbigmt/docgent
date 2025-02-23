package slack

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestConversationService_GetURI(t *testing.T) {
	tests := []struct {
		name   string
		handle ConversationHandle
		want   string
	}{
		{
			name: "スレッドの最初のメッセージの場合",
			handle: ConversationHandle{
				TeamID:                 "T123456",
				ChannelID:              "C789012",
				ThreadTimestamp:        "1234567890.123456",
				SourceMessageTimestamp: "1234567890.123456",
			},
			want: "https://app.slack.com/client/T123456/C789012/1234567890.123456",
		},
		{
			name: "スレッド内の返信メッセージの場合",
			handle: ConversationHandle{
				TeamID:                 "T123456",
				ChannelID:              "C789012",
				ThreadTimestamp:        "1234567890.123456",
				SourceMessageTimestamp: "1234567890.654321",
			},
			want: "https://app.slack.com/client/T123456/C789012/thread/C789012-1234567890.123456/1234567890.654321",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			service := &ConversationService{
				handle: tt.handle,
			}
			got := service.GetURI()
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestParseConversationURI(t *testing.T) {
	tests := []struct {
		name    string
		uri     string
		want    ConversationHandle
		wantErr bool
	}{
		{
			name: "スレッドの最初のメッセージのURIの場合",
			uri:  "https://app.slack.com/client/T123456/C789012/1234567890.123456",
			want: ConversationHandle{
				TeamID:                 "T123456",
				ChannelID:              "C789012",
				ThreadTimestamp:        "1234567890.123456",
				SourceMessageTimestamp: "1234567890.123456",
			},
			wantErr: false,
		},
		{
			name: "スレッド内の返信メッセージのURIの場合",
			uri:  "https://app.slack.com/client/T123456/C789012/thread/C789012-1234567890.123456/1234567890.654321",
			want: ConversationHandle{
				TeamID:                 "T123456",
				ChannelID:              "C789012",
				ThreadTimestamp:        "1234567890.123456",
				SourceMessageTimestamp: "1234567890.654321",
			},
			wantErr: false,
		},
		{
			name:    "不正なURIの場合",
			uri:     "https://invalid.url/T123456/C789012",
			want:    ConversationHandle{},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseConversationURI(tt.uri)
			if tt.wantErr {
				assert.Error(t, err)
				return
			}
			assert.NoError(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}
