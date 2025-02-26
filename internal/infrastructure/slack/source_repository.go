package slack

import (
	"context"
	"docgent/internal/domain/data"
	"fmt"
	"strings"

	"github.com/slack-go/slack"
)

type SourceRepository struct {
	slackAPI *API
}

func NewSourceRepository(slackAPI *API) *SourceRepository {
	return &SourceRepository{slackAPI: slackAPI}
}

func (r *SourceRepository) Match(uri *data.URI) bool {
	return uri.Host() == "app.slack.com"
}

func (r *SourceRepository) Find(ctx context.Context, uri *data.URI) (*data.Source, error) {
	ref, err := ParseConversationRef(uri)
	if err != nil {
		return nil, fmt.Errorf("failed to parse conversation ref: %w", err)
	}

	client := r.slackAPI.GetClient()
	messages, _, _, err := client.GetConversationReplies(&slack.GetConversationRepliesParameters{
		ChannelID: ref.ChannelID(),
		Timestamp: ref.ThreadTimestamp(),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get thread messages: %w", err)
	}

	var content strings.Builder
	content.WriteString(fmt.Sprintf("<conversation uri=%q>\n", uri))

	for _, message := range messages {
		// スレッド内の特定のメッセージが指定されている場合、そのメッセージにマークを付ける
		if message.Timestamp == ref.SourceMessageTimestamp() && ref.ThreadTimestamp() != ref.SourceMessageTimestamp() {
			content.WriteString(fmt.Sprintf("<message user=%q highlighted=\"true\">\n%s\n</message>\n", message.User, message.Text))
		} else {
			content.WriteString(fmt.Sprintf("<message user=%q>\n%s\n</message>\n", message.User, message.Text))
		}
	}

	content.WriteString("</conversation>")

	return data.NewSource(uri, content.String()), nil
}
