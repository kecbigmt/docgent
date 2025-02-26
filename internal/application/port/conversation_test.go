package port

import (
	"docgent/internal/domain/data"
	"strings"
	"testing"
)

func TestConversationHistory_ToXML(t *testing.T) {
	tests := []struct {
		name     string
		history  ConversationHistory
		expected []string // XMLの中に含まれるべき文字列
	}{
		{
			name: "通常の会話履歴",
			history: ConversationHistory{
				URI: data.NewURIUnsafe("https://app.slack.com/client/T00000000/C00000000/thread/T00000000-00000000"),
				Messages: []ConversationMessage{
					{
						Author:       "user1",
						Content:      "こんにちは",
						YouMentioned: false,
						IsYou:        false,
					},
					{
						Author:       "bot",
						Content:      "はい、こんにちは",
						YouMentioned: false,
						IsYou:        true,
					},
				},
			},
			expected: []string{
				`<conversation uri="https://app.slack.com/client/T00000000/C00000000/thread/T00000000-00000000">`,
				`<message author="user1">こんにちは</message>`,
				`<message author="bot" is_you="true">はい、こんにちは</message>`,
				`</conversation>`,
			},
		},
		{
			name: "メンション付きメッセージを含む会話",
			history: ConversationHistory{
				URI: data.NewURIUnsafe("https://app.slack.com/client/T00000000/C00000000/thread/T00000000-00000000"),
				Messages: []ConversationMessage{
					{
						Author:       "user2",
						Content:      "@bot 質問があります",
						YouMentioned: true,
						IsYou:        false,
					},
				},
			},
			expected: []string{
				`<conversation uri="https://app.slack.com/client/T00000000/C00000000/thread/T00000000-00000000">`,
				`<message author="user2" you_mentioned="true">@bot 質問があります</message>`,
				`</conversation>`,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.history.ToXML()

			// 期待される文字列が全て含まれているか確認
			for _, exp := range tt.expected {
				if !strings.Contains(result, exp) {
					t.Errorf("ToXML() の結果に %q が含まれていません\n結果: %s", exp, result)
				}
			}
		})
	}
}
